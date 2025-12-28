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
    image.pixels = (unsigned char *)malloc(image_size(&image));
    return image;
}

image_t image_clone(const image_t *image) {
    return image_create(image->width, image->height, image->bpp);
}

/**
 * Clear all pixel data in an image by setting every byte in the pixel buffer
 * to zero (regardless of image type).
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
    if (x < 0 || y < 0 || x >= (int)image->width || y >= (int)image->height) {
        return;
    }
    p = image->pixels + (x + y * image->width) * image->bpp;
    *r = *p++;
    *g = *p++;
    *b = *p++;
    *a = *p;
}

void apply_kernel(image_t *image, int size, int kernel[size][size], int divisor) {
    int x, y;
    image_t tmp;
    int limit = size/2;

    tmp = image_clone(image);

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
            image_putpixel(&tmp, x, y, r, g, b, 0);
        }
    }
    memcpy(image->pixels, tmp.pixels, image_size(image));
    free(tmp.pixels);
}

void filter_smooth_3x3_block(image_t *image) {
    int kernel[3][3] = {
        {0,1,0},
        {1,4,1},
        {0,1,0},
    };

    apply_kernel(image, 3, kernel, 9);
}

