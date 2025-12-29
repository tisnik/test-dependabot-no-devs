#include <math.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#define PI 3.1415927
#define EPSILON 0.01
#define DIST2(x1, y1, x2, y2) \
    (((x1) - (x2)) * ((x1) - (x2)) + ((y1) - (y2)) * ((y1) - (y2)))
#define MIN_FP_VALUE 1.0e-100

#define GRAYSCALE 1
#define RGB 3
#define RGBA 4

#define MAX(a, b) ((a) > (b) ? (a) : (b))
#define MIN(a, b) ((a) < (b) ? (a) : (b))

#define NULL_CHECK(value)       \
    if (value == NULL) {        \
        puts("NULL parameter"); \
        return;                 \
    }

/* ABI type definitions */

typedef struct {
    unsigned int   width;
    unsigned int   height;
    unsigned int   bpp;
    unsigned char *pixels;
} image_t;

/**
 * Compute the total size in bytes of an image's pixel buffer.
 *
 * @param image Pointer to the image whose buffer size will be computed.
 * @returns Total number of bytes required for the image's pixel buffer (width * height * bpp).
 */
size_t image_size(const image_t *image) {
    return image->width * image->height * image->bpp;
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
    image.width = width;
    image.height = height;
    image.bpp = bpp;
    /* callers must check that image.pixels != NULL */
    image.pixels = (unsigned char *)malloc(image_size(&image));
    return image;
}

/**
 * Create a new image with the same dimensions and bytes-per-pixel as the given image.
 *
 * @param image Source image to clone.
 * @returns A newly created image_t with the same width, height, and bpp as `image`; the pixel buffer is separately allocated and may be NULL if allocation fails.
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
 * to zero (regardles of image type).
 *
 * @param image Image whose pixel buffer will be cleared; must have a valid
 *              pixel buffer.
 */
void image_clear(image_t *image) {
    if (image == NULL || image->pixels == NULL) {
        return;
    }
    memset(image->pixels, 0x00, image_size(image));
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
 */
void image_putpixel(image_t *image, int x, int y, unsigned char r,
                    unsigned char g, unsigned char b, unsigned char a) {
    unsigned char *p;
    if (image == NULL || image->pixels == NULL) {
        return;
    }
    if (x < 0 || y < 0 || x >= (int)image->width || y >= (int)image->height) {
        return;
    }
    p = image->pixels + (x + y * image->width) * image->bpp;
    *p++ = r;
    *p++ = g;
    *p++ = b;
    if (image->bpp == RGBA) {
        *p = a;
    }
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
 */
void image_putpixel_max(image_t *image, int x, int y, unsigned char r, unsigned char g, unsigned char b, unsigned char a) {
    unsigned char *p;
    if (image == NULL || image->pixels == NULL) {
        return;
    }
    if (x < 0 || y < 0 || x >= (int)image->width || y >= (int)image->height) {
        return;
    }
    p = image->pixels + (x + y * image->width) * image->bpp;
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
 */
void image_getpixel(const image_t *image, int x, int y, unsigned char *r, unsigned char *g, unsigned char *b, unsigned char *a) {
    unsigned char *p;
    if (image == NULL || image->pixels == NULL || r == NULL || g == NULL || b == NULL || a == NULL) {
        return;
    }
    if (x < 0 || y < 0 || x >= (int)image->width || y >= (int)image->height) {
        return;
    }
    p = image->pixels + (x + y * image->width) * image->bpp;
    *r = *p++;
    *g = *p++;
    *b = *p++;

    if (image->bpp == RGBA) {
        *a = *p;
    } else {
        *a = 255; /* default opaque for non-RGBA images */
    }
}


/**
 * Apply a convolution kernel to the image, producing a filtered version in-place.
 *
 * Applies the provided size×size integer kernel to each pixel inside the image
 * (excluding a border of floor(size/2) pixels). For each processed pixel the
 * weighted sums of the R, G, B channels are computed, divided by `divisor`, and
 * written back into the image buffer; the alpha channel of written pixels is set to 0.
 * Border pixels that cannot be fully covered by the kernel are left unchanged.
 *
 * @param image   Image to be filtered; its pixel buffer is updated with the result.
 * @param size    Kernel dimension; must match both kernel array dimensions and be an odd positive integer.
 * @param kernel  2D integer kernel of dimensions [size][size]; kernel[row][col] is applied around each pixel.
 * @param divisor Value used to normalize the accumulated channel sums; must be non-zero.
 */
void apply_kernel(image_t *image, int size, int kernel[size][size], int divisor) {
    int x, y;
    image_t tmp;
    int limit = size/2;

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

    for (y=limit; y<tmp.height-limit; y++) {
        for (x=limit; x<tmp.width-limit; x++) {
            int r=0, g=0, b=0;
            int dx, dy;
            for (dy=-limit; dy<=limit; dy++) {
                for (dx=-limit; dx<=limit; dx++) {
                    unsigned char rr, gg, bb, aa;
                    image_getpixel(image, x+dx, y+dy, &rr, &gg, &bb, &aa);
                    r+=rr*kernel[dy+limit][dx+limit];
                    g+=gg*kernel[dy+limit][dx+limit];
                    b+=bb*kernel[dy+limit][dx+limit];
                }
            }
            r/=divisor;
            g/=divisor;
            b/=divisor;
            /* clamp to valid unsigned char range */
            r = (r < 0) ? 0 : (r > 255 ? 255 : r);
            g = (g < 0) ? 0 : (g > 255 ? 255 : g);
            b = (b < 0) ? 0 : (b > 255 ? 255 : b);
            image_putpixel(&tmp, x, y, r, g, b, 0);
        }
    }
    memcpy(image->pixels, tmp.pixels, image_size(image));
    free(tmp.pixels);
}

/**
 * Apply a 3×3 weighted smoothing filter to the given image in-place.
 *
 * Uses a 3×3 kernel with weights:
 *   1 1 1
 *   1 1 1
 *   1 1 1
 * and a divisor of 9 to perform a weighted average of each pixel's neighbourhood.
 *
 * @param image Image to be filtered; its pixel data is modified in-place.
 */
void filter_smooth_3x3_block(image_t *image) {
    static int kernel[3][3] = {
        {1,1,1},
        {1,1,1},
        {1,1,1},
    };

    apply_kernel(image, 3, kernel, 9);
}

void filter_smooth_3x3_gauss(image_t *image) {
    static int kernel[3][3] = {
        {1,2,1},
        {2,4,2},
        {1,2,1},
    };

    apply_kernel(image, 3, kernel, 16);
}

void filter_smooth_3x3_sharpen(image_t *image) {
    static int kernel[3][3] = {
        { 0,-1, 0},
        {-1, 5,-1},
        { 0,-1, 0},
    };

    apply_kernel(image, 3, kernel, 1);
}

void filter_smooth_3x3_edge_detection_1(image_t *image) {
    static int kernel[3][3] = {
        { 0,-1, 0},
        {-1, 4,-1},
        { 0,-1, 0},
    };

    apply_kernel(image, 3, kernel, 1);
}

void filter_smooth_3x3_edge_detection_2(image_t *image) {
    static int kernel[3][3] = {
        {-1,-1,-1},
        {-1, 8,-1},
        {-1,-1,-1},
    };

    apply_kernel(image, 3, kernel, 1);
}

void filter_smooth_3x3_edge_detection_3(image_t *image) {
    static int kernel[3][3] = {
        { 0, 1, 0},
        { 1,-4, 1},
        { 0, 1, 0},
    };

    apply_kernel(image, 3, kernel, 1);
}

void filter_smooth_3x3_horizontal_edge_detection(image_t *image) {
    static int kernel[3][3] = {
        {-1,-1,-1},
        { 0, 0, 0},
        { 1, 1, 1},
    };

    apply_kernel(image, 3, kernel, 1);
}

void filter_smooth_3x3_vertical_edge_detection(image_t *image) {
    static int kernel[3][3] = {
        {-1, 0, 1},
        {-1, 0, 1},
        {-1, 0, 1},
    };

    apply_kernel(image, 3, kernel, 1);
}

void filter_smooth_3x3_horizontal_sobel_operator(image_t *image) {
    static int kernel[3][3] = {
        {-1, 0, 1},
        {-2, 0, 2},
        {-1, 0, 1},
    };

    apply_kernel(image, 3, kernel, 1);
}

void filter_smooth_3x3_vertical_sobel_operator(image_t *image) {
    static int kernel[3][3] = {
        {-1,-2,-1},
        { 0, 0, 0},
        { 1, 2, 1},
    };

    apply_kernel(image, 3, kernel, 1);
}

void filter_smooth_3x3_laplacian(image_t *image) {
    static int kernel[3][3] = {
        { 0,-1, 0},
        {-1, 4,-1},
        { 0,-1, 0},
    };

    apply_kernel(image, 3, kernel, 1);
}

/**
 * Entry point for the program.
 * @returns 0 on successful execution.
 */
#ifndef NO_MAIN
/**
 * Program entry point that executes the specified rendering operations.
 *
 * @returns Exit code produced by the test sequence (e.g., `0` on success).
 */
int main(int argc, char **argv) {
    int result = 0;
    return result;
}
#endif

