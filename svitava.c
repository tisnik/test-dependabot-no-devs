/*

   Svitava: fractal renderer

   (C) Copyright 2024, 2025, 2026  Pavel Tisnovsky

   All rights reserved. This program and the accompanying materials
   are made available under the terms of the Eclipse Public License v1.0
   which accompanies this distribution, and is available at
   http://www.eclipse.org/legal/epl-v10.html

   Contributors:
       Pavel Tisnovsky

*/

/*
build as shared library:
    gcc -shared -Wl,-soname,svitava -o svitava.so -fPIC svitava.c

build as executable:
    gcc -lm -o svitava svitava.c
*/

/*
Overall structure:
------------------

Raster image filters implemented:
---------------------------------
filter_smooth_3x3_block
filter_smooth_3x3_gauss
filter_sharpen_3x3
filter_edge_detection_3x3_1
filter_edge_detection_3x3_2
filter_edge_detection_3x3_3
filter_horizontal_edge_detection_3x3
filter_vertical_edge_detection_3x3
filter_horizontal_sobel_operator_3x3
filter_vertical_sobel_operator_3x3

Image compositors:
--------------------
composite_interlace
composite_horizontal_interlace
composite_vertical_interlace
composite_blend

Image export operations:
------------------------
image_export_ppm_ascii
image_export_ppm_binary
image_export_bmp
image_export_tga
image_export_png

Image import operations:
------------------------

Renderers implemented:
----------------------

render_test_rgb_image
render_test_palette_image
render_mandelbrot
render_julia
render_mandelbrot_3
render_julia_3

*/

#include <stddef.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#ifdef SUPPORT_PNG
#include <png.h>
#endif

/* Image types */
#define GRAYSCALE 1
#define RGB 3
#define RGBA 4

/* Maximum image resolution */
#define MAX_WIDTH 8192
#define MAX_HEIGHT 8192

#define MAX(a, b) ((a) > (b) ? (a) : (b))
#define MIN(a, b) ((a) < (b) ? (a) : (b))

/* ABI type definitions */

/**
 * Structure that represents raster image of configurable resolution and bits
 * per pixel format.
 */
typedef struct {
    unsigned int   width;
    unsigned int   height;
    unsigned int   bpp;
    unsigned char *pixels;
} image_t;

enum error {
    OK,
    NULL_POINTER,
    NULL_IMAGE_POINTER,
    NULL_PIXELS_POINTER,
    NULL_COLOR_COMPONENT_POINTER,
    NULL_PALETTE_POINTER,
    INVALID_IMAGE_DIMENSION,
    INVALID_IMAGE_TYPE,
    INVALID_COORDINATES
};

/* Helper macros */
#define ENSURE_PROPER_IMAGE_STRUCTURE   \
    do {                                \
        if (image == NULL) {            \
            return NULL_IMAGE_POINTER;  \
        }                               \
        if (image->pixels == NULL) {    \
            return NULL_PIXELS_POINTER; \
        }                               \
    } while (0);

#define ENSURE_PROPER_PALETTE            \
    do {                                 \
        if (palette == NULL) {           \
            return NULL_PALETTE_POINTER; \
        }                                \
    } while (0);

/* Avoid division by zero */
#define ENSURE_NON_EMPTY_IMAGE                         \
    do {                                               \
        if (image->width == 0 || image->height == 0) { \
            return INVALID_IMAGE_DIMENSION;            \
        }                                              \
    } while (0);

/**
 * Compute the total size in bytes of an image's pixel buffer.
 *
 * @param image Pointer to the image whose buffer size will be computed.
 *
 * @returns Total number of bytes required for the image's pixel buffer
 *          (width * height * bpp).
 */
size_t image_size(const image_t *image) {
    if (image == NULL) {
        return 0;
    }
    /* cast to size_t before multiplication to prevent overflow */
    return (size_t)image->width * (size_t)image->height * (size_t)image->bpp;
}

/**
 * Create an image_t with the given width, height, and bytes-per-pixel,
 * allocating a pixel buffer.
 *
 * The returned image_t fields width, height, and bpp are initialized and
 * pixels points to a newly allocated buffer of size width * height * bpp. If
 * allocation fails, pixels will be NULL.
 *
 * @param width  Image width specified in pixels.
 * @param height Image height specified in pixels.
 * @param bpp    Bytes per pixel (bytes used to store a single pixel).
 *
 * @returns The initialized image_t; its `pixels` member points to the
 *          allocated buffer or NULL on allocation failure.
 */
image_t image_create(const unsigned int width, const unsigned int height, const unsigned int bpp) {
    image_t image;

    /* validate image size */
    if (width == 0 || height == 0 || width > MAX_WIDTH || height > MAX_HEIGHT) {
        image.width = 0;
        image.height = 0;
        image.bpp = 0;
        image.pixels = NULL;
        return image;
    }

    /* validate image type */
    if (bpp != GRAYSCALE && bpp != RGB && bpp != RGBA) {
        image.width = 0;
        image.height = 0;
        image.bpp = 0;
        image.pixels = NULL;
        return image;
    }

    /* initialize image */
    image.width = width;
    image.height = height;
    image.bpp = bpp;

    /* callers must check that image.pixels != NULL */
    image.pixels = (unsigned char *)malloc(image_size(&image));

    /* make sure the image will be 'zero value' when pixels are not allocated */
    if (image.pixels == NULL) {
        image.width = 0;
        image.height = 0;
        image.bpp = 0;
    }
    return image;
}

/**
 * Create a new image with the same dimensions and bytes-per-pixel as the given
 * image.
 *
 * If `image` is NULL or its pixel buffer is NULL, returns an image with
 * width=0, height=0, bpp=0 and pixels=NULL.
 *
 * @param image Source image to clone.
 *
 * @returns A newly created image_t with the same width, height, and bpp as
 * `image`; the pixel buffer is separately allocated and may be NULL if
 * allocation fails.
 */
image_t image_clone(const image_t *image) {
    image_t clone;
    if (image == NULL || image->pixels == NULL) {
        clone.width = 0;
        clone.height = 0;
        clone.bpp = 0;
        clone.pixels = NULL;
        return clone;
    }
    clone = image_create(image->width, image->height, image->bpp);
    if (clone.pixels != NULL) {
        memcpy(clone.pixels, image->pixels, image_size(image));
    }
    return clone;
}

/**
 * Clear all pixel data in an image by setting every byte in the pixel buffer
 * to zero (regardless of image type).
 *
 * @param image Image whose pixel buffer will be cleared; must have a valid
 *              pixel buffer.
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          OK otherwise
 */
int image_clear(image_t *image) {

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE

    memset(image->pixels, 0x00, image_size(image));
    return OK;
}

/**
 * Set the RGBA color of the pixel at the specified (x, y) coordinates in the image.
 *
 * If (x, y) lies outside the image bounds the function has no effect.
 *
 * @param image Pointer to the image whose pixel will be updated.
 * @param x Horizontal pixel coordinate (0 is left).
 * @param y Vertical pixel coordinate (0 is top).
 * @param r Red component (0–255).
 * @param g Green component (0–255).
 * @param b Blue component (0–255).
 * @param a Alpha component (0–255).
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          INVALID_COORDINATES when pixel coordinate(s) are out of range
 *          OK otherwise
 */
int image_putpixel(image_t *image, int x, int y, unsigned char r,
                   unsigned char g, unsigned char b, unsigned char a) {
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE

    if (x < 0 || y < 0 || x >= (int)image->width || y >= (int)image->height) {
        return INVALID_COORDINATES;
    }
    p = image->pixels + (x + y * image->width) * image->bpp;
    if (image->bpp == GRAYSCALE) {
        /* convert to grayscale using integer approximation of standard weights */
        /* uses integer arithmetic with coefficients scaled by 256 (77≈0.299×256, 150≈0.587×256, 29≈0.114×256) */
        *p = (unsigned char)((77 * r + 150 * g + 29 * b) >> 8);
    } else {
        *p++ = r;
        *p++ = g;
        *p++ = b;
        if (image->bpp == RGBA) {
            *p = a;
        }
    }
    return OK;
}

/**
 * Update the pixel at (x, y) by replacing each color channel with the greater of
 * the existing channel and the provided value; the alpha channel is written
 * unconditionally.
 *
 * If (x, y) is outside the image bounds, the function does nothing.
 *
 * @param image Target image.
 * @param x X coordinate of the pixel.
 * @param y Y coordinate of the pixel.
 * @param r Red component candidate; pixel's red becomes `max(current, r)`.
 * @param g Green component candidate; pixel's green becomes `max(current, g)`.
 * @param b Blue component candidate; pixel's blue becomes `max(current, b)`.
 * @param a Alpha component to write (overwrites existing alpha).
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          INVALID_COORDINATES when pixel coordinate(s) are out of range
 *          OK otherwise
 */
int image_putpixel_max(image_t *image, int x, int y, unsigned char r, unsigned char g, unsigned char b, unsigned char a) {
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE

    if (x < 0 || y < 0 || x >= (int)image->width || y >= (int)image->height) {
        return INVALID_COORDINATES;
    }
    p = image->pixels + (x + y * image->width) * image->bpp;
    if (image->bpp == GRAYSCALE) {
        /* convert to grayscale using integer approximation of standard weights */
        /* uses integer arithmetic with coefficients scaled by 256 (77≈0.299×256, 150≈0.587×256, 29≈0.114×256) */
        unsigned char gray = (unsigned char)((77 * r + 150 * g + 29 * b) >> 8);
        if (*p < gray) {
            *p = gray;
        }
        return OK;
    }
    /* image type is RGB or RGBA */
    if (*p < r) {
        *p = r;
    }
    p++;
    if (*p < g) {
        *p = g;
    }
    p++;
    if (*p < b) {
        *p = b;
    }
    if (image->bpp == RGBA) {
        p++;
        *p = a;
    }
    return OK;
}

/**
 * Retrieve the RGBA components of the pixel at (x, y).
 *
 * If (x, y) is outside the image bounds the function returns without modifying
 * the output pointers. The output pointers must be non-NULL when (x, y) is
 * inside bounds.
 *
 * @param image Pointer to the source image.
 * @param x X coordinate of the pixel.
 * @param y Y coordinate of the pixel.
 * @param r Pointer to receive the red component (0–255).
 * @param g Pointer to receive the green component (0–255).
 * @param b Pointer to receive the blue component (0–255).
 * @param a Pointer to receive the alpha component (0–255).
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          NULL_COLOR_COMPONENT_POINTER when pointer to color component is NULL,
 *          INVALID_COORDINATES when pixel coordinate(s) are out of range
 *          OK otherwise
 */
int image_getpixel(const image_t *image, int x, int y, unsigned char *r, unsigned char *g, unsigned char *b, unsigned char *a) {
    const unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE

    if (x < 0 || y < 0 || x >= (int)image->width || y >= (int)image->height) {
        return INVALID_COORDINATES;
    }
    if (r == NULL || g == NULL || b == NULL || a == NULL) {
        return NULL_COLOR_COMPONENT_POINTER;
    }
    p = image->pixels + (x + y * image->width) * image->bpp;

    if (image->bpp == GRAYSCALE) {
        /* for grayscale, replicate the single value to all RGB channels */
        *r = *g = *b = *p;
        *a = 255; /* grayscale images are always opaque */
        return OK;
    }

    *r = *p++;
    *g = *p++;
    *b = *p++;
    if (image->bpp == RGBA) {
        *a = *p;
    } else {
        *a = 255; /* default opaque for non-RGBA images */
    }
    return OK;
}

/**
 * Draws a horizontal line between two x coordinates at a given y using the specified RGBA color.
 *
 * The line includes both endpoints; the order of `x1` and `x2` does not matter. Pixels that lie
 * outside the image bounds are ignored.
 *
 * @param image Target image to draw into.
 * @param x1 One end x coordinate of the line.
 * @param x2 Other end x coordinate of the line.
 * @param y Y coordinate of the line.
 * @param r Red component (0–255).
 * @param g Green component (0–255).
 * @param b Blue component (0–255).
 * @param a Alpha component (0–255).
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          INVALID_COORDINATES when pixel coordinate(s) are out of range
 *          OK otherwise
 */
int image_hline(image_t *image, int x1, int x2, int y, unsigned char r, unsigned char g, unsigned char b, unsigned char a) {
    int x, fromX = MIN(x1, x2), toX = MAX(x1, x2);

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE

    if (fromX < 0 || y < 0 || toX >= (int)image->width || y >= (int)image->height) {
        return INVALID_COORDINATES;
    }
    for (x = fromX; x <= toX; x++) {
        /* all checks are performed internally */
        /* TODO: fast putpixel function w/o any checks */
        int result = image_putpixel(image, x, y, r, g, b, a);
        if (result != OK) {
            return result;
        }
    }
    return OK;
}

/**
 * Draws a vertical line at column x between y1 and y2 inclusive using the specified RGBA color.
 *
 * The line includes both endpoints; the order of `y1` and `y2` does not matter. Pixels that lie
 * outside the image bounds are ignored.
 *
 * @param image Target image.
 * @param x X coordinate (column) where the line is drawn.
 * @param y1 One end Y coordinate of the line.
 * @param y2 Other end Y coordinate of the line.
 * @param r Red component (0-255).
 * @param g Green component (0-255).
 * @param b Blue component (0-255).
 * @param a Alpha component (0-255).
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          INVALID_COORDINATES when pixel coordinate(s) are out of range
 *          OK otherwise
 */
int image_vline(image_t *image, int x, int y1, int y2, unsigned char r, unsigned char g, unsigned char b, unsigned char a) {
    int y, fromY = MIN(y1, y2), toY = MAX(y1, y2);

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE

    if (fromY < 0 || x < 0 || toY >= (int)image->height || x >= (int)image->width) {
        return INVALID_COORDINATES;
    }
    for (y = fromY; y <= toY; y++) {
        /* all checks are performed internally */
        /* TODO: fast putpixel function w/o any checks */
        int result = image_putpixel(image, x, y, r, g, b, a);
        if (result != OK) {
            return result;
        }
    }
    return OK;
}

/**
 * Draws a straight line between two pixel coordinates using an integer rasterization algorithm.
 *
 * The line includes both endpoint pixels and writes the specified RGBA color to each covered pixel.
 * Pixels that lie outside the image bounds are ignored.
 *
 * @param image Target image to draw into.
 * @param x1 X coordinate of the start point (in pixels).
 * @param y1 Y coordinate of the start point (in pixels).
 * @param x2 X coordinate of the end point (in pixels).
 * @param y2 Y coordinate of the end point (in pixels).
 * @param r Red component of the color (0-255).
 * @param g Green component of the color (0-255).
 * @param b Blue component of the color (0-255).
 * @param a Alpha component of the color (0-255).
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          INVALID_COORDINATES when pixel coordinate(s) are out of range
 *          OK otherwise
 */
int image_line(image_t *image, int x1, int y1, int x2, int y2, unsigned char r, unsigned char g, unsigned char b, unsigned char a) {
    int dx = abs(x2 - x1), sx = x1 < x2 ? 1 : -1;
    int dy = abs(y2 - y1), sy = y1 < y2 ? 1 : -1;
    int err = (dx > dy ? dx : -dy) / 2, e2;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE

    while (1) {
        /* all checks are performed internally */
        /* TODO: fast putpixel function w/o any checks */
        int result = image_putpixel(image, x1, y1, r, g, b, a);
        if (result != OK) {
            return result;
        }
        /* we reached endpoint */
        if (x1 == x2 && y1 == y2) {
            return OK;
        }
        e2 = err;
        if (e2 > -dx) {
            err -= dy;
            x1 += sx;
        }
        if (e2 < dy) {
            err += dx;
            y1 += sy;
        }
    }
}

/**
 * Draws an anti-aliased straight line between two points into an RGBA image.
 *
 * The line is rasterized with sub-pixel intensity distribution so adjacent pixels
 * receive proportionally scaled color components for smoothing. Color components
 * are applied using the image's per-pixel maximum blending semantics (brightest
 * component wins); alpha is written as provided.
 *
 * @param image Target image buffer (RGBA) to draw into.
 * @param x1 X coordinate of the line start.
 * @param y1 Y coordinate of the line start.
 * @param x2 X coordinate of the line end.
 * @param y2 Y coordinate of the line end.
 * @param r Red component (0–255) of the line color.
 * @param g Green component (0–255) of the line color.
 * @param b Blue component (0–255) of the line color.
 * @param a Alpha component (0–255) of the line color.
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          INVALID_COORDINATES when pixel coordinate(s) are out of range
 *          INVALID_IMAGE_TYPE for images that are not of type RGBA
 *          OK otherwise
 */
int image_line_aa(image_t *image, int x1, int y1, int x2, int y2, unsigned char r, unsigned char g, unsigned char b, unsigned char a) {
    int    dx = x2 - x1;
    int    dy = y2 - y1;
    double s, p, e = 0.0;
    int    x, y, xdelta, ydelta, xpdelta, ypdelta, xp, yp;
    int    i, imin, imax;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE

    /* anti-aliasing requires RGBA for proper blending */
    if (image->bpp != RGBA) {
        return INVALID_IMAGE_TYPE;
    }

    /* strict vertical line does not have to be anti-aliased */
    if (x1 == x2) {
        return image_vline(image, x1, y1, y2, r, g, b, a);
    }

    /* strict horizontal line does not have to be anti-aliased */
    if (y1 == y2) {
        return image_hline(image, x1, x2, y1, r, g, b, a);
    }

    if (x1 > x2) {
        x1 = x1 ^ x2;
        x2 = x1 ^ x2;
        x1 = x1 ^ x2;
        y1 = y1 ^ y2;
        y2 = y1 ^ y2;
        y1 = y1 ^ y2;
    }

    /* iterate along the dominant (longer) axis */
    if (abs(dx) > abs(dy)) {
        s = (double)dy / (double)dx; /* slope: rise over run */
        imin = x1;
        imax = x2;
        x = x1;
        y = y1;
        xdelta = 1;
        ydelta = 0;
        xpdelta = 0;
        xp = 0;
        if (y2 > y1) {
            ypdelta = 1;
            yp = 1;
        } else {
            s = -s;
            ypdelta = -1;
            yp = -1;
        }
    } else {
        s = (double)dx / (double)dy; /* slope: run over rise */
        xdelta = 0;
        ydelta = 1;
        ypdelta = 0;
        yp = 0;
        if (y2 > y1) {
            imin = y1;
            imax = y2;
            x = x1;
            y = y1;
            xpdelta = 1;
            xp = 1;
        } else {
            s = -s;
            imin = y2;
            imax = y1;
            x = x2;
            y = y2;
            xpdelta = -1;
            xp = -1;
        }
    }
    /* p: sub-pixel step scaled to [0, 256) range for intensity calculation */
    p = s * 256.0;
    for (i = imin; i <= imax; i++) {
        int c1, c2;
        c1 = (int)e;
        if (c1 > 255) {
            c1 = 255;
        }
        c2 = 255 - c1;
        /* TODO: fast putpixel function w/o any checks */
        image_putpixel_max(image, x + xp, y + yp, (r * c1) / 255, (g * c1) / 255, (b * c1) / 255, a);
        image_putpixel_max(image, x, y, (r * c2) / 255, (g * c2) / 255, (b * c2) / 255, a);
        e = e + p;
        x += xdelta;
        y += ydelta;
        if (e >= 256.0) {
            e -= 256.0;
            x += xpdelta;
            y += ypdelta;
        }
    }
    return OK;
}

/**
 * Apply a convolution kernel to the image, producing a filtered version in-place.
 *
 * Applies the provided size×size integer kernel to each pixel inside the image
 * (excluding a border of floor(size/2) pixels). For each processed pixel the
 * weighted sums of the R, G, B channels are computed, divided by `divisor`, and
 * written back into the image buffer; the alpha channel of written pixels is
 * set to 255 (fully opaque). Border pixels that cannot be fully covered by the
 * kernel are left unchanged.
 *
 * @param image   Image to be filtered; its pixel buffer is updated with the result.
 * @param size    Kernel dimension; must match both kernel array dimensions and be an odd positive integer.
 * @param kernel  2D integer kernel of dimensions [size][size]; kernel[row][col] is applied around each pixel.
 * @param divisor Value used to normalize the accumulated channel sums; must be non-zero.
 */
void apply_kernel(image_t *image, int size, int kernel[size][size], int divisor) {
    int     x, y;
    image_t tmp;
    int     limit = size / 2;

    if (image == NULL || image->pixels == NULL) {
        return;
    }
    /* size must be odd positive number */
    if (size <= 0 || size % 2 == 0 || divisor == 0) {
        return;
    }

    tmp = image_clone(image);
    if (tmp.pixels == NULL) {
        return; /* allocation failed */
    }

    for (y = limit; y < (int)tmp.height - limit; y++) {
        for (x = limit; x < (int)tmp.width - limit; x++) {
            int r = 0, g = 0, b = 0;
            int dx, dy;
            for (dy = -limit; dy <= limit; dy++) {
                for (dx = -limit; dx <= limit; dx++) {
                    unsigned char rr, gg, bb, aa;
                    image_getpixel(image, x + dx, y + dy, &rr, &gg, &bb, &aa);
                    r += rr * kernel[dy + limit][dx + limit];
                    g += gg * kernel[dy + limit][dx + limit];
                    b += bb * kernel[dy + limit][dx + limit];
                }
            }
            r /= divisor;
            g /= divisor;
            b /= divisor;
            /* clamp to valid unsigned char range */
            r = (r < 0) ? 0 : (r > 255 ? 255 : r);
            g = (g < 0) ? 0 : (g > 255 ? 255 : g);
            b = (b < 0) ? 0 : (b > 255 ? 255 : b);
            image_putpixel(&tmp, x, y, r, g, b, 255);
        }
    }
    memcpy(image->pixels, tmp.pixels, image_size(image));
    free(tmp.pixels);
}

/**
 * Apply a 3×3 weighted smoothing filter to the given image in-place.
 *
 * Uses a 3×3 kernel with weights:
 *   [ 1, 1, 1 ]
 *   [ 1, 1, 1 ]
 *   [ 1, 1, 1 ]
 * and a divisor of 9 to perform a weighted average of each pixel's neighbourhood.
 *
 * @param image Image to be filtered; the pixel buffer is modified in-place. If `image` or its pixel buffer is NULL, no action is taken.
 */
void filter_smooth_3x3_block(image_t *image) {
    static int kernel[3][3] = {
        {1, 1, 1},
        {1, 1, 1},
        {1, 1, 1},
    };

    apply_kernel(image, 3, kernel, 9);
}

/**
 * Apply a 3×3 Gaussian-like smoothing filter to the provided image in-place.
 *
 * Uses the 3×3 kernel with weights:
 *   [ 1, 2, 1 ]
 *   [ 2, 4, 2 ]
 *   [ 1, 2, 1 ]
 * and a divisor of 16 to perform smoothing.
 *
 * @param image Image to be filtered; the pixel buffer is modified in-place. If `image` or its pixel buffer is NULL, no action is taken.
 */
void filter_smooth_3x3_gauss(image_t *image) {
    static int kernel[3][3] = {
        {1, 2, 1},
        {2, 4, 2},
        {1, 2, 1},
    };

    apply_kernel(image, 3, kernel, 16);
}

/**
 * Apply a 3×3 sharpening filter to the image in place.
 *
 * Uses the 3×3 kernel with weights:
 *   [ 0, -1,  0 ]
 *   [-1,  5, -1 ]
 *   [ 0, -1,  0 ]
 *
 * @param image Image to be filtered; the pixel buffer is modified in-place. If `image` or its pixel buffer is NULL, no action is taken.
 */
void filter_sharpen_3x3(image_t *image) {
    static int kernel[3][3] = {
        {0, -1, 0},
        {-1, 5, -1},
        {0, -1, 0},
    };

    apply_kernel(image, 3, kernel, 1);
}

/**
 * Apply a 3×3 edge-detection filter (4-neighbor Laplacian kernel) to the image in-place.
 *
 * This filter highlights regions of rapid intensity change (edges) by computing the
 * second derivative approximation. Edges appear as bright pixels; negative values
 * are clamped to zero.
 *
 * The kernel applied is:
 *   [ 0, -1,  0 ]
 *   [-1,  4, -1 ]
 *   [ 0, -1,  0 ]
 *
 * @param image Image to be filtered; the pixel buffer is modified in-place. If `image` or its pixel buffer is NULL, no action is taken.
 */
void filter_edge_detection_3x3_1(image_t *image) {
    static int kernel[3][3] = {
        {0, -1, 0},
        {-1, 4, -1},
        {0, -1, 0},
    };

    apply_kernel(image, 3, kernel, 1);
}

/**
 * Apply a 3×3 edge-detection filter (8-neighbor Laplacian kernel) to the image in-place.
 *
 * This filter highlights regions of rapid intensity change (edges) using all eight
 * neighboring pixels. Edges appear as bright pixels; negative values are clamped to zero.
 *
 * The kernel applied is:
 *   [-1, -1, -1 ]
 *   [-1,  8, -1 ]
 *   [-1, -1, -1 ]
 *
 * @param image Image to be filtered; the pixel buffer is modified in-place. If `image` or its pixel buffer is NULL, no action is taken.
 */
void filter_edge_detection_3x3_2(image_t *image) {
    static int kernel[3][3] = {
        {-1, -1, -1},
        {-1, 8, -1},
        {-1, -1, -1},
    };

    apply_kernel(image, 3, kernel, 1);
}

/**
 * Apply a 3×3 Laplacian-like edge-detection filter to the provided image in-place.
 *
 * This filter uses the inverted polarity of filter_edge_detection_3x3_1, highlighting
 * edges where the center pixel is darker than its neighbors. Edges appear as bright
 * pixels; negative values are clamped to zero.
 *
 * The kernel applied is:
 *   [ 0,  1,  0 ]
 *   [ 1, -4,  1 ]
 *   [ 0,  1,  0 ]
 *
 * @param image Image to be filtered; the pixel buffer is modified in-place. If `image` or its pixel buffer is NULL, no action is taken.
 */
void filter_edge_detection_3x3_3(image_t *image) {
    static int kernel[3][3] = {
        {0, 1, 0},
        {1, -4, 1},
        {0, 1, 0},
    };

    apply_kernel(image, 3, kernel, 1);
}

/**
 * Apply a 3×3 horizontal edge-detection filter to an image in-place.
 *
 * The kernel applied is:
 *   [-1, -1, -1]
 *   [ 0,  0,  0]
 *   [ 1,  1,  1]
 *
 * @param image Image to be filtered; the pixel buffer is modified in-place. If `image` or its pixel buffer is NULL, no action is taken.
 */
void filter_horizontal_edge_detection_3x3(image_t *image) {
    static int kernel[3][3] = {
        {-1, -1, -1},
        {0, 0, 0},
        {1, 1, 1},
    };

    apply_kernel(image, 3, kernel, 1);
}

/**
 * Apply a 3×3 vertical edge-detection filter to the image in-place.
 *
 * The filter highlights vertical edges by convolving the image with a 3×3
 * vertical edge-detection kernel.
 *
 * The kernel applied is:
 *   [-1,  0,  1]
 *   [-1,  0,  1]
 *   [-1,  0,  1]
 *
 * @param image Image to be filtered; the pixel buffer is modified in-place. If `image` or its pixel buffer is NULL, no action is taken.
 */
void filter_vertical_edge_detection_3x3(image_t *image) {
    static int kernel[3][3] = {
        {-1, 0, 1},
        {-1, 0, 1},
        {-1, 0, 1},
    };

    apply_kernel(image, 3, kernel, 1);
}

/**
 * Apply the 3×3 horizontal Sobel operator to the given image, modifying pixels in-place.
 *
 * The kernel applied is:
 *   [-1,  0,  1]
 *   [-2,  0,  2]
 *   [-1,  0,  1]
 *
 * @param image Image to be filtered; the pixel buffer is modified in-place. If `image` or its pixel buffer is NULL, no action is taken.
 */
void filter_horizontal_sobel_operator_3x3(image_t *image) {
    static int kernel[3][3] = {
        {-1, 0, 1},
        {-2, 0, 2},
        {-1, 0, 1},
    };

    apply_kernel(image, 3, kernel, 1);
}

/**
 * Apply the 3×3 vertical Sobel operator to the given image, modifying pixels in-place.
 *
 * The kernel applied is:
 *   [-1, -2, -1]
 *   [ 0,  0,  0]
 *   [ 1,  2,  1]
 *
 * @param image Image to be filtered; the pixel buffer is modified in-place. If `image` or its pixel buffer is NULL, no action is taken.
 */
void filter_vertical_sobel_operator_3x3(image_t *image) {
    static int kernel[3][3] = {
        {-1, -2, -1},
        {0, 0, 0},
        {1, 2, 1},
    };

    apply_kernel(image, 3, kernel, 1);
}

/**
 * Validate that three images are suitable for composition operations.
 * Returns 1 if valid, 0 otherwise.
 */
static int validate_composition_inputs(const image_t *src1, const image_t *src2, const image_t *dest) {
    /* validate inputs */
    if (!src1 || !src2 || !dest) {
        return 0;
    }

    /* pixel buffers must exist */
    if (!src1->pixels || !src2->pixels || !dest->pixels) {
        return 0;
    }

    /* validate supported formats */
    if ((src1->bpp != GRAYSCALE && src1->bpp != RGB && src1->bpp != RGBA) ||
        (src2->bpp != GRAYSCALE && src2->bpp != RGB && src2->bpp != RGBA) ||
        (dest->bpp != GRAYSCALE && dest->bpp != RGB && dest->bpp != RGBA)) {
        return 0;
    }

    /* ensure all images have the same dimensions */
    if (src1->width != src2->width || src1->height != src2->height ||
        src1->width != dest->width || src1->height != dest->height) {
        return 0;
    }

    return 1;
}

/*
 * Interleave pixels from two source images into a destination image using a horizontal pattern.
 *
 * For each pixel position (x,y), selects the pixel from `src1` when x is odd and from `src2` when x is even, then writes that RGBA pixel into `dest`.
 *
 * @param src1 Source image providing pixels for odd columns; must have the same dimensions as `src2` and `dest`.
 * @param src2 Source image providing pixels for even columns; must have the same dimensions as `src1` and `dest`.
 * @param dest Destination image receiving the interleaved pixels; must have the same dimensions as `src1` and `src2`.
 */
void composite_horizontal_interlace(const image_t *src1, const image_t *src2, image_t *dest) {
    unsigned int i, j;

    if (!validate_composition_inputs(src1, src2, dest)) {
        return;
    }

    for (j = 0; j < src1->height; j++) {
        for (i = 0; i < src1->width; i++) {
            unsigned char r, g, b, a;
            int           which = i % 2;
            if (which) {
                image_getpixel(src1, i, j, &r, &g, &b, &a);
            } else {
                image_getpixel(src2, i, j, &r, &g, &b, &a);
            }
            image_putpixel(dest, i, j, r, g, b, a);
        }
    }
}

/**
 * Interleave pixels from two source images into a destination image using a vertical pattern.
 *
 * For each pixel position, pixels on odd-numbered rows are taken from `src1`
 * and pixels on even-numbered rows are taken from `src2`; the selected RGBA
 * values are written into `dest` at the same coordinates.
 *
 * @param src1 Source image supplying pixels for odd rows.
 * @param src2 Source image supplying pixels for even rows.
 * @param dest Destination image receiving the interleaved pixels.
 */
void composite_vertical_interlace(const image_t *src1, const image_t *src2, image_t *dest) {
    unsigned int i, j;

    if (!validate_composition_inputs(src1, src2, dest)) {
        return;
    }

    for (j = 0; j < src1->height; j++) {
        for (i = 0; i < src1->width; i++) {
            unsigned char r, g, b, a;
            int           which = j % 2;
            if (which) {
                image_getpixel(src1, i, j, &r, &g, &b, &a);
            } else {
                image_getpixel(src2, i, j, &r, &g, &b, &a);
            }
            image_putpixel(dest, i, j, r, g, b, a);
        }
    }
}

/**
 * Compose a destination image by selecting pixels from two sources in a checkerboard pattern.
 *
 * For each coordinate (i, j), the destination receives the pixel from `src1` when (i + j) is odd;
 * otherwise the pixel is taken from `src2`.
 *
 * @param src1 Source image providing pixels for one set of checkerboard positions.
 * @param src2 Source image providing pixels for the alternating checkerboard positions.
 * @param dest Destination image where the composed pixels are written. Must have the same dimensions
 *             as `src1` and `src2`.
 */
void composite_checkberboard_interlace(const image_t *src1, const image_t *src2, image_t *dest) {
    unsigned int i, j;

    if (!validate_composition_inputs(src1, src2, dest)) {
        return;
    }

    for (j = 0; j < src1->height; j++) {
        for (i = 0; i < src1->width; i++) {
            unsigned char r, g, b, a;
            int           which = (i % 2) ^ (j % 2);
            if (which) {
                image_getpixel(src1, i, j, &r, &g, &b, &a);
            } else {
                image_getpixel(src2, i, j, &r, &g, &b, &a);
            }
            image_putpixel(dest, i, j, r, g, b, a);
        }
    }
}

/**
 * Blend two source images into a destination by averaging corresponding RGBA channels.
 *
 * Each destination pixel is written with the per-channel average of the two source pixels:
 * channel = (channel_src1 + channel_src2) >> 1 (integer division by 2).
 *
 * @param src1 First source image; its width and height determine the processed area.
 * @param src2 Second source image; pixels are read at the same coordinates as src1.
 * @param dest Destination image that will be written with the blended pixels.
 */
void composite_blend(const image_t *src1, const image_t *src2, image_t *dest) {
    unsigned int i, j;

    if (!validate_composition_inputs(src1, src2, dest)) {
        return;
    }

    for (j = 0; j < src1->height; j++) {
        for (i = 0; i < src1->width; i++) {
            unsigned char r1, g1, b1, a1;
            unsigned char r2, g2, b2, a2;
            image_getpixel(src1, i, j, &r1, &g1, &b1, &a1);
            image_getpixel(src2, i, j, &r2, &g2, &b2, &a2);
            image_putpixel(dest, i, j, (r1 + r2) >> 1, (g1 + g2) >> 1, (b1 + b2) >> 1, (a1 + a2) >> 1);
        }
    }
}

/**
 * Writes pixel data to a file stream in ASCII PPM (P3) format.
 *
 * The output image is written from bottom to top, with each pixel's RGB values
 * output as text. Assumes the pixel buffer uses 4 bytes per pixel, with the
 * fourth byte ignored.
 */
void ppm_write_ascii_to_stream(unsigned int width, unsigned int height,
                               unsigned char *pixels, FILE *fout) {
    int            x, y;
    unsigned char  r, g, b;
    unsigned char *p = pixels;

    /* header */
    fprintf(fout, "P3 %d %d 255\n", width, height);

    /* pixel array */
    for (y = height - 1; y >= 0; y--) {
        /* TODO: fix bottom-to-up ordering */
        for (x = 0; x < width; x++) {
            r = *p++;
            g = *p++;
            b = *p++;
            p++;
            fprintf(fout, "%d %d %d\n", r, g, b);
        }
    }
}

void ppm_write_binary_to_stream(unsigned int width, unsigned int height,
                                unsigned char *pixels, FILE *fout) {
    int            x, y;
    unsigned char  rgb[3];
    unsigned char *p = pixels;

    /* header */
    fprintf(fout, "P6 %d %d 255\n", width, height);

    /* pixel array */
    for (y = height - 1; y >= 0; y--) {
        /* TODO: fix bottom-to-up ordering */
        for (x = 0; x < width; x++) {
            memcpy(rgb, p, 3);
            p += 4;
            fwrite(rgb, sizeof(rgb), 1, fout);
        }
    }
}

/**
 * Writes pixel data to a file in ASCII PPM (P3) format.
 *
 * @param width Image width in pixels.
 * @param height Image height in pixels.
 * @param pixels Pointer to the pixel buffer (assumed 4 bytes per pixel, RGB in
 * first 3 bytes).
 * @param file_name Name of the output file.
 * @return 0 on success, -1 on failure to open or close the file.
 */
int image_export_ppm_ascii(unsigned int width, unsigned int height,
                           unsigned char *pixels, const char *file_name) {
    FILE *fout;

    fout = fopen(file_name, "wb");
    if (!fout) {
        return -1;
    }

    ppm_write_ascii_to_stream(width, height, pixels, fout);

    if (fclose(fout) == EOF) {
        return -1;
    }
    return 0;
}

/**
 * Writes pixel data to a file in binary PPM (P6) format.
 *
 * @param width Image width in pixels.
 * @param height Image height in pixels.
 * @param pixels Pointer to the pixel buffer (assumed 4 bytes per pixel, RGB in
 * first 3 bytes).
 * @param file_name Name of the output file.
 * @return 0 on success, -1 on failure to open or close the file.
 */
int image_export_ppm_binary(unsigned int width, unsigned int height,
                            unsigned char *pixels, const char *file_name) {
    FILE *fout;

    fout = fopen(file_name, "wb");
    if (!fout) {
        return -1;
    }

    ppm_write_binary_to_stream(width, height, pixels, fout);

    if (fclose(fout) == EOF) {
        return -1;
    }
    return 0;
}

/**
 * Writes a pixel buffer to a BMP file with 24 bits per pixel.
 *
 * The pixel data is written in bottom-up order, with each pixel's RGB channels
 * reordered as BGR. The function assumes the input buffer uses 4 bytes per
 * pixel, with the fourth byte ignored. Returns 0 on success, or 1 if the file
 * cannot be opened.
 *
 * @param width Width of the image in pixels.
 * @param height Height of the image in pixels.
 * @param pixels Pointer to the pixel buffer (4 bytes per pixel, only RGB used).
 * @param file_name Name of the output BMP file.
 * @return 0 on success, 1 on failure to open the file.
 */
int image_export_bmp(unsigned int width, unsigned int height, unsigned char *pixels,
                     const char *file_name) {
    unsigned char bmp_header[] = {
        /* BMP header structure: */
        0x42, 0x4d,             /* magic number */
        0x46, 0x00, 0x00, 0x00, /* header size in bytes (set to 70, placeholder and unused) */
        0x00, 0x00,             /* unused */
        0x00, 0x00,             /* unused */
        0x36, 0x00, 0x00, 0x00, /* 54 bytes - offset to data */
        0x28, 0x00, 0x00, 0x00, /* 40 bytes - bytes in DIB header */
        0x00, 0x00, 0x00, 0x00, /* width of bitmap */
        0x00, 0x00, 0x00, 0x00, /* height of bitmap */
        0x01, 0x0,              /* 1 pixel plane */
        0x18, 0x00,             /* 24 bpp */
        0x00, 0x00, 0x00, 0x00, /* no compression */
        0x00, 0x00, 0x00, 0x00, /* size of pixel array */
        0x13, 0x0b, 0x00, 0x00, /* 2835 pixels/meter */
        0x13, 0x0b, 0x00, 0x00, /* 2835 pixels/meter */
        0x00, 0x00, 0x00, 0x00, /* color palette */
        0x00, 0x00, 0x00, 0x00, /* important colors */
    };
    FILE *fout;
    int   x, y;

    /* BMP rows must be padded to 4-byte boundaries */
    unsigned int  row_padding = (4 - (width * 3) % 4) % 4;
    unsigned char pad[3] = {0, 0, 0};

    if (pixels == NULL || file_name == NULL) {
        return -1;
    }

    /* initialize BMP header */
    bmp_header[18] = width & 0xff;
    bmp_header[19] = (width >> 8) & 0xff;
    bmp_header[20] = (width >> 16) & 0xff;
    bmp_header[21] = (width >> 24) & 0xff;
    bmp_header[22] = height & 0xff;
    bmp_header[23] = (height >> 8) & 0xff;
    bmp_header[24] = (height >> 16) & 0xff;
    bmp_header[25] = (height >> 24) & 0xff;

    fout = fopen(file_name, "wb");
    if (!fout) {
        return -1;
    }

    /* write BMP header */
    fwrite(bmp_header, sizeof(bmp_header), 1, fout);

    /* write the whole pixel array into BMP file */
    for (y = height - 1; y >= 0; y--) {
        /* pointer to the 1st pixel on scan line */
        unsigned char *p = pixels + y * width * 4;
        for (x = 0; x < width; x++) {
            /* swap RGB color components as required by file format */
            fwrite(p + 2, 1, 1, fout);
            fwrite(p + 1, 1, 1, fout);
            fwrite(p, 1, 1, fout);
            /* move to next pixel on scan line */
            p += 4;
        }
        /* write padding bytes to align row to 4-byte boundary */
        if (row_padding > 0) {
            fwrite(pad, 1, row_padding, fout);
        }
    }
    fclose(fout);
    return 0;
}

static const unsigned char true_color_tga_header[] = {
    0x00,                   /* without image ID */
    0x00,                   /* color map type: without palette */
    0x02,                   /* uncompressed true color image */
    0x00, 0x00,             /* start of color palette (it is not used) */
    0x00, 0x00,             /* length of color palette (it is not used) */
    0x00,                   /* bits per palette entry */
    0x00, 0x00, 0x00, 0x00, /* image coordinates */
    0x00, 0x00,             /* image width */
    0x00, 0x00,             /* image height */
    0x18,                   /* bits per pixel = 24 */
    0x20                    /* picture orientation: top-left origin */
};

/**
 * Export raw RGBA pixel data as an uncompressed 24-bit TGA image (top-left origin).
 *
 * Expects pixels in 4-byte RGBA order; writes width*height pixels as BGR triplets (alpha byte is ignored).
 * @param width Image width in pixels.
 * @param height Image height in pixels.
 * @param pixels Pointer to pixel buffer containing width*height*4 bytes in RGBA order.
 * @param file_name Destination file path for the TGA output.
 * @returns 0 on success, 1 if `pixels` is NULL, -1 on file open/write/close failure.
 */
int image_export_tga(unsigned int width, unsigned int height,
                     const unsigned char *pixels, const char *file_name) {
    FILE                *fout;
    const unsigned char *p = pixels;
    unsigned char        header[sizeof true_color_tga_header];
    int                  i;

    if (pixels == NULL || file_name == NULL) {
        return -1;
    }

    /* prepare a local header copy to avoid mutating the global template */
    memcpy(header, true_color_tga_header, sizeof header);

    fout = fopen(file_name, "wb");
    if (!fout) {
        return -1;
    }
    /* image size is specified in TGA header */
    header[12] = (width) & 0xff;
    header[13] = (width) >> 8;
    header[14] = (height) & 0xff;
    header[15] = (height) >> 8;

    /* write TGA header */
    if (fwrite(header, sizeof header, 1, fout) != 1) {
        fclose(fout);
        return -1;
    }

    /* write the whole pixel array into TGA file */
    for (i = 0; i < width * height; i++) {
        /* swap RGB to BGR */
        unsigned char bgr[3];
        bgr[0] = p[2];
        bgr[1] = p[1];
        bgr[2] = p[0];
        /* write RGB, but not alpha */
        if (fwrite(bgr, sizeof bgr, 1, fout) != 1) {
            fclose(fout);
            return -1;
        }
        p += 4; /* skip alpha */
    }

    if (fclose(fout) == EOF) {
        return -1;
    }
    return 0;
}

#ifdef SUPPORT_PNG
/**
 * Export raw RGBA pixel data as an PNG.
 *
 * Expects pixels in 4-byte RGBA order
 * @param width Image width in pixels.
 * @param height Image height in pixels.
 * @param pixels Pointer to pixel buffer containing width*height*4 bytes in RGBA order.
 * @param file_name Destination file path for the PNG output.
 * @returns 0 on success, 1 if `pixels` is NULL, -1 on file open/write/close failure.
 */
int image_export_png(unsigned int width, unsigned int height,
                     const unsigned char *pixels, const char *file_name) {
    FILE                *fout;
    const unsigned char *p = pixels;
    int code = 0;
    int scanline;
    png_structp png_ptr = NULL;
    png_infop info_ptr = NULL;

    char *title = NULL;

    if (pixels == NULL || file_name == NULL) {
        return -1;
    }

    /* open file for writing in binary mode */
    fout = fopen(file_name, "wb");
    if (fout == NULL)
    {
        fprintf(stderr, "Could not open file %s for writing\n", file_name);
        code = 1;
        goto FINALISE;
    }

    /* initialize write structure */
    png_ptr = png_create_write_struct(PNG_LIBPNG_VER_STRING, NULL, NULL, NULL);
    if (png_ptr == NULL)
    {
        fprintf(stderr, "Could not allocate write struct\n");
        code = 1;
        goto FINALISE;
    }

    /* initialize info structure */
    info_ptr = png_create_info_struct(png_ptr);
    if (info_ptr == NULL)
    {
        fprintf(stderr, "Could not allocate info struct\n");
        code = 1;
        goto FINALISE;
    }

    /* setup Exception handling */
    if (setjmp(png_jmpbuf(png_ptr)))
    {
        fprintf(stderr, "Error during png creation\n");
        code = 1;
        goto FINALISE;
    }

    png_init_io(png_ptr, fout);

    /* write header (8 bit colour depth) */
    png_set_IHDR(png_ptr, info_ptr, width, height,
            8, PNG_COLOR_TYPE_RGBA, PNG_INTERLACE_NONE,
            PNG_COMPRESSION_TYPE_BASE, PNG_FILTER_TYPE_BASE);

    /* set the title */
    if (title != NULL)
    {
        png_text title_text;
        title_text.compression = PNG_TEXT_COMPRESSION_NONE;
        title_text.key = "Title";
        title_text.text = title;
        png_set_text(png_ptr, info_ptr, &title_text, 1);
    }

    png_write_info(png_ptr, info_ptr);

    /* write image data */
    for (scanline=0 ; scanline<height; scanline++)
    {
        png_write_row(png_ptr, p);
        /* TODO: not only RGBA! */
        p += width * 4;
    }

    /* end write */
    png_write_end(png_ptr, NULL);

FINALISE:
    if (fout != NULL)
    {
        fclose(fout);
    }
    if (info_ptr != NULL)
    {
        png_free_data(png_ptr, info_ptr, PNG_FREE_ALL, -1);
    }
    if (png_ptr != NULL)
    {
        png_destroy_write_struct(&png_ptr, (png_infopp)NULL);
    }

    return code;
}
#endif

/**
 * Writes an RGB color from the palette at the specified index into the pixel
 * buffer and advances the pixel pointer by 4 bytes.
 *
 * @returns none
 */
void putpixel(unsigned char **pixel, const unsigned char *palette,
              int color_index) {
    int            color_offset = color_index * 3;
    unsigned char *pal = (unsigned char *)(palette + color_offset);

    *(*pixel)++ = *pal++;
    *(*pixel)++ = *pal++;
    *(*pixel)++ = *pal;
    *(*pixel)++ = 255;
}

/**
 * Fills the pixel buffer with a test RGB image where the red channel is set to
 * the x-coordinate, the green channel is set to a fixed value, and the blue
 * channel is set to the y-coordinate.
 *
 * The pixel buffer is assumed to use 4 bytes per pixel, with the fourth byte
 * unused or as padding.
 *
 * @param green Value to assign to the green channel for all pixels.
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          NULL_PALETTE_POINTER when the palette is NULL
 *          INVALID_IMAGE_TYPE for images that are not of type RGBA
 *          OK otherwise
 */
int render_test_rgb_image(const image_t *image, const unsigned char *palette,
                          unsigned char green) {
    unsigned int   i, j;
    unsigned char *p;
    unsigned int   div_x, div_y;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    p = image->pixels;

    div_x = image->width / 256;
    div_y = image->height / 256;

    for (j = 0; j < image->height; j++) {
        for (i = 0; i < image->width; i++) {
            *p++ = i / div_x;
            *p++ = green;
            *p++ = j / div_y;
            p++;
        }
    }
    return OK;
}

/**
 * Fills the pixel buffer with a test image using colors from the palette
 * indexed by the x-coordinate.
 *
 * Each pixel in the image is assigned a color from the palette based on its
 * horizontal position, creating vertical color bands.
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          NULL_PALETTE_POINTER when the palette is NULL
 *          INVALID_IMAGE_TYPE for images that are not of type RGBA
 *          OK otherwise
 */
int render_test_palette_image(const image_t *image, const unsigned char *palette) {
    unsigned int   i, j;
    unsigned char *p;
    unsigned int   div_x;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    /* avoid empty images */
    if (image->width == 0 || image->height == 0) {
        return INVALID_IMAGE_DIMENSION;
    }

    p = image->pixels;
    div_x = image->width / 256;

    for (j = 0; j < image->height; j++) {
        for (i = 0; i < image->width; i++) {
            int color = i / div_x;
            putpixel(&p, palette, color);
        }
    }
    return OK;
}

/**
 * Renders the Mandelbrot set fractal into a pixel buffer using a specified
 * color palette.
 *
 * Iterates over each pixel, maps it to a point in the complex plane, and
 * computes the escape time for the Mandelbrot set. The iteration count
 * determines the color index used from the palette.
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          NULL_PALETTE_POINTER when the palette is NULL
 *          INVALID_IMAGE_TYPE for images that are not of type RGBA
 *          INVALID_IMAGE_DIMENSION if image has zero width/height
 *          OK otherwise
 */
int render_mandelbrot(const image_t *image, const unsigned char *palette,
                      double zx0, double zy0, int maxiter) {
    int            x, y;
    double         cx, cy;
    double         xmin = -2.0, ymin = -1.5, xmax = 1.0, ymax = 1.5;
    double         step_x;
    double         step_y;
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    /* putpixel() writes 4 bytes per pixel, so only RGBA images are supported */
    if (image->bpp != RGBA) {
        return INVALID_IMAGE_TYPE;
    }

    p = image->pixels;
    step_x = (xmax - xmin) / image->width;
    step_y = (ymax - ymin) / image->height;

    cy = ymin;
    for (y = 0; y < image->height; y++) {
        cx = xmin;
        for (x = 0; x < image->width; x++) {
            double       zx = zx0;
            double       zy = zy0;
            unsigned int i = 0;
            while (i < maxiter) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                zy = 2.0 * zx * zy + cy;
                zx = zx2 - zy2 + cx;
                i++;
            }
            putpixel(&p, palette, i % 256);
            cx += step_x;
        }
        cy += step_y;
    }
    return OK;
}

/**
 * Renders a Julia set fractal image into the provided pixel buffer.
 *
 * Maps each pixel to a point in the complex plane and iterates the function z =
 * z^2 + c, where c is specified by `cx` and `cy`. Colors are assigned based on
 * the number of iterations before escape, using the provided palette.
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          NULL_PALETTE_POINTER when the palette is NULL
 *          INVALID_IMAGE_TYPE for images that are not of type RGBA
 *          INVALID_IMAGE_DIMENSION if image has zero width/height
 *          OK otherwise
 */
int render_julia(const image_t *image, const unsigned char *palette,
                 double cx, double cy, int maxiter) {
    int            x, y;
    double         zx0, zy0;
    double         xmin = -1.5, ymin = -1.5, xmax = 1.5, ymax = 1.5;
    double         step_x;
    double         step_y;
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    /* putpixel() writes 4 bytes per pixel, so only RGBA images are supported */
    if (image->bpp != RGBA) {
        return INVALID_IMAGE_TYPE;
    }

    p = image->pixels;
    step_x = (xmax - xmin) / image->width;
    step_y = (ymax - ymin) / image->height;

    zy0 = ymin;
    for (y = 0; y < image->height; y++) {
        zx0 = xmin;
        for (x = 0; x < image->width; x++) {
            double       zx = zx0;
            double       zy = zy0;
            unsigned int i = 0;
            while (i < maxiter) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                zy = 2.0 * zx * zy + cy;
                zx = zx2 - zy2 + cx;
                i++;
            }
            putpixel(&p, palette, i % 256);
            zx0 += step_x;
        }
        zy0 += step_y;
    }
    return OK;
}

/**
 * Renders a cubic Mandelbrot fractal (z = z³ + c) into the pixel buffer.
 *
 * Each pixel is mapped to a point in the complex plane, and the cubic
 * Mandelbrot iteration is performed up to `maxiter` times. The number of
 * iterations before escape determines the color index from the palette.
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          NULL_PALETTE_POINTER when the palette is NULL
 *          INVALID_IMAGE_TYPE for images that are not of type RGBA
 *          INVALID_IMAGE_DIMENSION if image has zero width/height
 *          OK otherwise
 */
int render_mandelbrot_3(const image_t *image, const unsigned char *palette,
                        double zx0, double zy0, int maxiter) {
    int            x, y;
    double         cx, cy;
    double         xmin = -1.5, ymin = -1.5, xmax = 1.5, ymax = 1.5;
    double         step_x;
    double         step_y;
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    /* putpixel() writes 4 bytes per pixel, so only RGBA images are supported */
    if (image->bpp != RGBA) {
        return INVALID_IMAGE_TYPE;
    }

    p = image->pixels;
    step_x = (xmax - xmin) / image->width;
    step_y = (ymax - ymin) / image->height;

    cy = ymin;
    for (y = 0; y < image->height; y++) {
        cx = xmin;
        for (x = 0; x < image->width; x++) {
            double       zx = zx0;
            double       zy = zy0;
            unsigned int i = 0;
            while (i < maxiter) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                double zx3 = zx2 * zx;
                double zy3 = zy2 * zy;
                double zxn, zyn;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                zxn = zx3 - 3.0 * zx * zy2 + cx;
                zyn = -zy3 + 3.0 * zx2 * zy + cy;
                zx = zxn;
                zy = zyn;
                i++;
            }
            putpixel(&p, palette, i % 256);
            cx += step_x;
        }
        cy += step_y;
    }
    return OK;
}

/**
 * Renders a cubic Julia set fractal image using the formula z = z³ + c.
 *
 * Each pixel is mapped to a point in the complex plane, and the cubic Julia
 * iteration is performed up to `maxiter` times or until the escape radius is
 * exceeded. The number of iterations before escape determines the color, which
 * is selected from the palette.
 *
 * @param cx Real part of the constant complex parameter c.
 * @param cy Imaginary part of the constant complex parameter c.
 * @param maxiter Maximum number of iterations for the escape test.
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          NULL_PALETTE_POINTER when the palette is NULL
 *          INVALID_IMAGE_TYPE for images that are not of type RGBA
 *          INVALID_IMAGE_DIMENSION if image has zero width/height
 *          OK otherwise
 */
int render_julia_3(const image_t *image, const unsigned char *palette,
                   double cx, double cy, int maxiter) {
    int            x, y;
    double         zx0, zy0;
    double         step_x;
    double         step_y;
    double         xmin = -1.5, ymin = -1.5, xmax = 1.5, ymax = 1.5;
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    /* putpixel() writes 4 bytes per pixel, so only RGBA images are supported */
    if (image->bpp != RGBA) {
        return INVALID_IMAGE_TYPE;
    }

    p = image->pixels;
    step_x = (xmax - xmin) / image->width;
    step_y = (ymax - ymin) / image->height;

    zy0 = ymin;
    for (y = 0; y < image->height; y++) {
        zx0 = xmin;
        for (x = 0; x < image->width; x++) {
            double       zx = zx0;
            double       zy = zy0;
            unsigned int i = 0;
            while (i < maxiter) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                double zx3 = zx2 * zx;
                double zy3 = zy2 * zy;
                double zxn, zyn;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                zxn = zx3 - 3.0 * zx * zy2 + cx;
                zyn = -zy3 + 3.0 * zx2 * zy + cy;
                zx = zxn;
                zy = zyn;
                i++;
            }
            putpixel(&p, palette, i % 256);
            zx0 += step_x;
        }
        zy0 += step_y;
    }
    return OK;
}

/**
 * Renders a quartic Mandelbrot set fractal image into a pixel buffer.
 *
 * Iterates the function z = z⁴ + c for each pixel mapped to the complex plane,
 * coloring pixels based on the number of iterations before escape, using the
 * provided palette. The fractal is rendered over the region [-1.5, 1.5] x
 * [-1.5, 1.5] in the complex plane.
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          NULL_PALETTE_POINTER when the palette is NULL
 *          INVALID_IMAGE_TYPE for images that are not of type RGBA
 *          INVALID_IMAGE_DIMENSION if image has zero width/height
 *          OK otherwise
 */
int render_mandelbrot_4(const image_t *image, const unsigned char *palette,
                         double zx0, double zy0, int maxiter) {
    int            x, y;
    double         cx, cy;
    double         xmin = -1.5, ymin = -1.5, xmax = 1.5, ymax = 1.5;
    double         step_x;
    double         step_y;
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    /* putpixel() writes 4 bytes per pixel, so only RGBA images are supported */
    if (image->bpp != RGBA) {
        return INVALID_IMAGE_TYPE;
    }

    p = image->pixels;
    step_x = (xmax - xmin) / image->width;
    step_y = (ymax - ymin) / image->height;

    cy = ymin;
    for (y = 0; y < image->height; y++) {
        cx = xmin;
        for (x = 0; x < image->width; x++) {
            double       zx = zx0;
            double       zy = zy0;
            unsigned int i = 0;
            while (i < maxiter) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                double zx4, zy4;
                double zxn, zyn;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                zxn = zx2 - zy2;
                zyn = 2.0 * zx * zy;
                zx4 = zxn * zxn;
                zy4 = zyn * zyn;
                zy = 2.0 * zxn * zyn + cy;
                zx = zx4 - zy4 + cx;
                i++;
            }
            putpixel(&p, palette, i % 256);
            cx += step_x;
        }
        cy += step_y;
    }
    return OK;
}

/**
 * Renders a quartic Julia set fractal image using the formula z = z⁴ + c.
 *
 * Iterates the quartic Julia set for each pixel, mapping image coordinates to
 * the complex plane in the range [-1.5, 1.5] for both axes. The constant
 * complex parameter c is specified by `cx` and `cy`. Each pixel is colored
 * based on the number of iterations before escape, using the provided palette.
 *
 * @param cx Real part of the constant parameter c for the Julia set.
 * @param cy Imaginary part of the constant parameter c for the Julia set.
 * @param maxiter Maximum number of iterations for the escape condition.
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          NULL_PALETTE_POINTER when the palette is NULL
 *          INVALID_IMAGE_TYPE for images that are not of type RGBA
 *          INVALID_IMAGE_DIMENSION if image has zero width/height
 *          OK otherwise
 */
int render_julia_4(const image_t *image, const unsigned char *palette,
                    double cx, double cy, int maxiter) {
    int            x, y;
    double         zx0, zy0;
    double         step_x;
    double         step_y;
    double         xmin = -1.5, ymin = -1.5, xmax = 1.5, ymax = 1.5;
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    /* putpixel() writes 4 bytes per pixel, so only RGBA images are supported */
    if (image->bpp != RGBA) {
        return INVALID_IMAGE_TYPE;
    }

    p = image->pixels;
    step_x = (xmax - xmin) / image->width;
    step_y = (ymax - ymin) / image->height;

    zy0 = ymin;
    for (y = 0; y < image->height; y++) {
        zx0 = xmin;
        for (x = 0; x < image->width; x++) {
            double       zx = zx0;
            double       zy = zy0;
            unsigned int i = 0;
            while (i < maxiter) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                double zx4, zy4;
                double zxn, zyn;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                zxn = zx2 - zy2;
                zyn = 2.0 * zx * zy;
                zx4 = zxn * zxn;
                zy4 = zyn * zyn;
                zy = 2.0 * zxn * zyn + cy;
                zx = zx4 - zy4 + cx;
                i++;
            }
            putpixel(&p, palette, i % 256);
            zx0 += step_x;
        }
        zy0 += step_y;
    }
    return OK;
}

/**
 * Renders the Barnsley Mandelbrot-type fractal (variant m1) into a pixel buffer.
 *
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          NULL_PALETTE_POINTER when the palette is NULL
 *          INVALID_IMAGE_TYPE for images that are not of type RGBA
 *          INVALID_IMAGE_DIMENSION if image has zero width/height
 *          OK otherwise
 */
int render_barnsley_m1(const image_t *image, const unsigned char *palette,
                        double zx0, double zy0, int maxiter) {
    int            x, y;
    double         cx, cy;
    double         xmin = -2.0, ymin = -2.0, xmax = 2.0, ymax = 2.0;
    double         step_x;
    double         step_y;
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    /* putpixel() writes 4 bytes per pixel, so only RGBA images are supported */
    if (image->bpp != RGBA) {
        return INVALID_IMAGE_TYPE;
    }

    p = image->pixels;
    step_x = (xmax - xmin) / image->width;
    step_y = (ymax - ymin) / image->height;

    cy = ymin;
    for (y = 0; y < image->height; y++) {
        cx = xmin;
        for (x = 0; x < image->width; x++) {
            double       zx = zx0;
            double       zy = zy0;
            unsigned int i = 0;
            while (i < maxiter) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                double zxn, zyn;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                if (zx >= 0) {
                    zxn = zx * cx - zy * cy - cx;
                    zyn = zx * cy + zy * cx - cy;
                } else {
                    zxn = zx * cx - zy * cy + cx;
                    zyn = zx * cy + zy * cx + cy;
                }
                zx = zxn;
                zy = zyn;
                i++;
            }
            putpixel(&p, palette, i % 256);
            cx += step_x;
        }
        cy += step_y;
    }
    return OK;
}

/**
 * Renders a Barnsley Julia fractal image using a piecewise complex function.
 *
 * Maps each pixel to a point in the complex plane and iterates a piecewise
 * function based on the sign of the real part, using parameters `cx` and `cy`.
 * Colors are assigned from the palette based on the number of iterations before
 * escape or reaching `maxiter`.
 */
int render_barnsley_j1(const image_t *image, const unsigned char *palette,
                        double cx, double cy, int maxiter) {
    int            x, y;
    double         zx0, zy0;
    double         xmin = -2.0, ymin = -2.0, xmax = 2.0, ymax = 2.0;
    double         step_x;
    double         step_y;
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    /* putpixel() writes 4 bytes per pixel, so only RGBA images are supported */
    if (image->bpp != RGBA) {
        return INVALID_IMAGE_TYPE;
    }

    p = image->pixels;
    step_x = (xmax - xmin) / image->width;
    step_y = (ymax - ymin) / image->height;

    zy0 = ymin;
    for (y = 0; y < image->height; y++) {
        zx0 = xmin;
        for (x = 0; x < image->width; x++) {
            double       zx = zx0;
            double       zy = zy0;
            unsigned int i = 0;
            while (i < maxiter) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                double zxn, zyn;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                if (zx >= 0) {
                    zxn = zx * cx - zy * cy - cx;
                    zyn = zx * cy + zy * cx - cy;
                } else {
                    zxn = zx * cx - zy * cy + cx;
                    zyn = zx * cy + zy * cx + cy;
                }
                zx = zxn;
                zy = zyn;
                i++;
            }
            putpixel(&p, palette, i % 256);
            zx0 += step_x;
        }
        zy0 += step_y;
    }
    return OK;
}

/**
 * Renders the Barnsley Mandelbrot-type fractal (variant m2) into a pixel buffer.
 *
 * The fractal is generated by iterating a piecewise function on the complex
 * plane for each pixel, with the function branch determined by the sign of (zx
 * * cy + zy * cx). The number of iterations before escape (or reaching maxiter)
 * determines the color index for each pixel.
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          NULL_PALETTE_POINTER when the palette is NULL
 *          INVALID_IMAGE_TYPE for images that are not of type RGBA
 *          INVALID_IMAGE_DIMENSION if image has zero width/height
 *          OK otherwise
 */
int render_barnsley_m2(const image_t *image, const unsigned char *palette,
                        double zx0, double zy0, int maxiter) {
    int            x, y;
    double         cx, cy;
    double         xmin = -2.0, ymin = -2.0, xmax = 2.0, ymax = 2.0;
    double         step_x;
    double         step_y;
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    /* putpixel() writes 4 bytes per pixel, so only RGBA images are supported */
    if (image->bpp != RGBA) {
        return INVALID_IMAGE_TYPE;
    }

    p = image->pixels;
    step_x = (xmax - xmin) / image->width;
    step_y = (ymax - ymin) / image->height;

    cy = ymin;
    for (y = 0; y < image->height; y++) {
        cx = xmin;
        for (x = 0; x < image->width; x++) {
            double       zx = zx0;
            double       zy = zy0;
            unsigned int i = 0;
            while (i < maxiter) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                double zxn, zyn;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                if (zx * cy + zy * cx >= 0) {
                    zxn = zx * cx - zy * cy - cx;
                    zyn = zx * cy + zy * cx - cy;
                } else {
                    zxn = zx * cx - zy * cy + cx;
                    zyn = zx * cy + zy * cx + cy;
                }
                zx = zxn;
                zy = zyn;
                i++;
            }
            putpixel(&p, palette, i % 256);
            cx += step_x;
        }
        cy += step_y;
    }
    return OK;
}

/**
 * Renders a Barnsley Julia set fractal variant using a piecewise complex
 * function.
 *
 * For each pixel, maps its coordinates to the complex plane and iterates a
 * piecewise function based on the sign of a linear combination of zx and zy,
 * with parameters cx and cy. Colors are assigned from the palette based on the
 * number of iterations before escape or reaching maxiter.
 */
int render_barnsley_j2(const image_t *image, const unsigned char *palette,
                        double cx, double cy, int maxiter) {
    int            x, y;
    double         zx0, zy0;
    double         xmin = -2.0, ymin = -2.0, xmax = 2.0, ymax = 2.0;
    double         step_x;
    double         step_y;
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    /* putpixel() writes 4 bytes per pixel, so only RGBA images are supported */
    if (image->bpp != RGBA) {
        return INVALID_IMAGE_TYPE;
    }

    p = image->pixels;
    step_x = (xmax - xmin) / image->width;
    step_y = (ymax - ymin) / image->height;

    zy0 = ymin;
    for (y = 0; y < image->height; y++) {
        zx0 = xmin;
        for (x = 0; x < image->width; x++) {
            double       zx = zx0;
            double       zy = zy0;
            unsigned int i = 0;
            while (i < maxiter) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                double zxn, zyn;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                if (zx * cy + zy * cx >= 0) {
                    zxn = zx * cx - zy * cy - cx;
                    zyn = zx * cy + zy * cx - cy;
                } else {
                    zxn = zx * cx - zy * cy + cx;
                    zyn = zx * cy + zy * cx + cy;
                }
                zx = zxn;
                zy = zyn;
                i++;
            }
            putpixel(&p, palette, i % 256);
            zx0 += step_x;
        }
        zy0 += step_y;
    }
    return OK;
}

/**
 * Renders the Barnsley Mandelbrot-type fractal (variant m3) into a pixel buffer.
 *
 * Iterates a piecewise quadratic function over the complex plane for each
 * pixel, with the function's behavior depending on the sign of the real part of
 * z. Colors are assigned based on the number of iterations before escape, up to
 * `maxiter`.
 *
 * @returns NULL_IMAGE_POINTER when the input image is NULL
 *          NULL_PIXELS_POINTER when pixels are not allocated,
 *          NULL_PALETTE_POINTER when the palette is NULL
 *          INVALID_IMAGE_TYPE for images that are not of type RGBA
 *          INVALID_IMAGE_DIMENSION if image has zero width/height
 *          OK otherwise
 */
int render_barnsley_m3(const image_t *image, const unsigned char *palette,
                        double zx0, double zy0, int maxiter) {
    int            x, y;
    double         cx, cy;
    double         xmin = -2.0, ymin = -2.0, xmax = 2.0, ymax = 2.0;
    double         step_x;
    double         step_y;
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    /* putpixel() writes 4 bytes per pixel, so only RGBA images are supported */
    if (image->bpp != RGBA) {
        return INVALID_IMAGE_TYPE;
    }

    p = image->pixels;
    step_x = (xmax - xmin) / image->width;
    step_y = (ymax - ymin) / image->height;

    cy = ymin;
    for (y = 0; y < image->height; y++) {
        cx = xmin;
        for (x = 0; x < image->width; x++) {
            double       zx = zx0;
            double       zy = zy0;
            unsigned int i = 0;
            while (i < maxiter) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                double zxn, zyn;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                if (zx > 0) {
                    zxn = zx2 - zy2 - 1;
                    zyn = 2.0 * zx * zy;
                } else {
                    zxn = zx2 - zy2 - 1 + cx * zx;
                    zyn = 2.0 * zx * zy + cy * zx;
                }
                zx = zxn;
                zy = zyn;
                i++;
            }
            putpixel(&p, palette, i % 256);
            cx += step_x;
        }
        cy += step_y;
    }
    return OK;
}

/**
 * Renders a Barnsley Julia-type fractal variant using a piecewise quadratic
 * iteration.
 *
 * For each pixel, maps coordinates to the complex plane and iterates a
 * piecewise function:
 * - If the real part is positive, applies z = z^2 - 1.
 * - If the real part is non-positive, applies z = z^2 - 1 + c*z, where c = cx +
 * i*cy. Iteration stops when the magnitude squared exceeds 4 or the maximum
 * number of iterations is reached. The number of iterations determines the
 * color index for each pixel.
 *
 * @param width Image width in pixels.
 * @param height Image height in pixels.
 * @param palette Pointer to the color palette (array of RGB triples).
 * @param pixels Output pixel buffer (4 bytes per pixel, RGB in first 3 bytes).
 * @param cx Real part of the Julia parameter c.
 * @param cy Imaginary part of the Julia parameter c.
 * @param maxiter Maximum number of iterations per pixel.
 */
int render_barnsley_j3(const image_t *image, const unsigned char *palette,
                        double cx, double cy, int maxiter) {
    int            x, y;
    double         zx0, zy0;
    double         xmin = -2.0, ymin = -2.0, xmax = 2.0, ymax = 2.0;
    double         step_x;
    double         step_y;
    unsigned char *p;

    /* Only properly-created images are accepted */
    ENSURE_PROPER_IMAGE_STRUCTURE
    ENSURE_NON_EMPTY_IMAGE
    ENSURE_PROPER_PALETTE

    /* putpixel() writes 4 bytes per pixel, so only RGBA images are supported */
    if (image->bpp != RGBA) {
        return INVALID_IMAGE_TYPE;
    }

    p = image->pixels;
    step_x = (xmax - xmin) / image->width;
    step_y = (ymax - ymin) / image->height;

    zy0 = ymin;
    for (y = 0; y < image->height; y++) {
        zx0 = xmin;
        for (x = 0; x < image->width; x++) {
            double       zx = zx0;
            double       zy = zy0;
            unsigned int i = 0;
            while (i < maxiter) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                double zxn, zyn;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                if (zx > 0) {
                    zxn = zx2 - zy2 - 1;
                    zyn = 2.0 * zx * zy;
                } else {
                    zxn = zx2 - zy2 - 1 + cx * zx;
                    zyn = 2.0 * zx * zy + cy * zx;
                }
                zx = zxn;
                zy = zyn;
                i++;
            }
            putpixel(&p, palette, i % 256);
            zx0 += step_x;
        }
        zy0 += step_y;
    }
    return OK;
}

#ifndef NO_MAIN
int main(int argc, char **argv) {
#define WIDTH 512
#define HEIGHT 512
    unsigned char *palette = (unsigned char *)malloc(256 * 3);
    image_t        image = image_create(WIDTH, HEIGHT, RGBA);
    int            i;

    for (i = 0; i <= 255; i++) {
        palette[i * 3] = i * 2;
        palette[i * 3 + 1] = i * 3;
        palette[i * 3 + 2] = i * 5;
    }

    /*
    image_clear(&image);
    render_test_rgb_image(&image, palette, 128);
    image_export_bmp(WIDTH, HEIGHT, image.pixels, "test_rgb.bmp");

    image_clear(&image);
    render_test_palette_image(&image, palette);
    image_export_bmp(WIDTH, HEIGHT, image.pixels, "test_palette.bmp");

    image_clear(&image);
    render_mandelbrot(&image, palette, 0.0, 0.0, 255);
    image_export_bmp(WIDTH, HEIGHT, image.pixels, "mandelbrot.bmp");

    image_clear(&image);
    render_julia(&image, palette, -0.207190825000000012496, 0.676656624999999999983, 255);
    image_export_bmp(WIDTH, HEIGHT, image.pixels, "julia.bmp");

    image_clear(&image);
    render_mandelbrot_3(&image, palette, 0.0, 0.0, 255);
    image_export_bmp(WIDTH, HEIGHT, image.pixels, "mandelbrot_3.bmp");

    image_clear(&image);
    render_julia_3(&image, palette, 0.12890625, -0.796875, 1000);
    image_export_bmp(WIDTH, HEIGHT, image.pixels, "julia_3.bmp");

    image_clear(&image);
    render_mandelbrot_4(&image, palette, 0.0, 0.0, 255);
    image_export_bmp(WIDTH, HEIGHT, image.pixels, "mandelbrot_4.bmp");

    image_clear(&image);
    render_julia_4(&image, palette, 0.375, -0.97265625, 1000);
    image_export_ppm_binary(WIDTH, HEIGHT, image.pixels, "julia_4.ppm");
*/
    image_clear(&image);
    render_barnsley_m1(&image, palette, 0, 0, 1000);
    image_export_ppm_binary(WIDTH, HEIGHT, image.pixels, "barnsley_m1.ppm");

    image_clear(&image);
    render_barnsley_j1(&image, palette, 0.4, 1.5, 1000);
    image_export_ppm_binary(WIDTH, HEIGHT, image.pixels, "barnsley_j1.ppm");

    image_clear(&image);
    render_barnsley_m2(&image, palette, 0, 0, 1000);
    image_export_ppm_binary(WIDTH, HEIGHT, image.pixels, "barnsley_m2.ppm");

    image_clear(&image);
    render_barnsley_j2(&image, palette, 1.109375, 0.421875, 1000);
    image_export_ppm_binary(WIDTH, HEIGHT, image.pixels, "barnsley_j2.ppm");

    image_clear(&image);
    render_barnsley_m3(&image, palette, 0, 0, 1000);
    image_export_ppm_binary(WIDTH, HEIGHT, image.pixels, "barnsley_m3.ppm");

    image_clear(&image);
    render_barnsley_j3(&image, palette, -0.09375, 0.453125, 1000);
    image_export_ppm_binary(WIDTH, HEIGHT, image.pixels, "barnsley_j3.ppm");

    return 0;
}
#endif
