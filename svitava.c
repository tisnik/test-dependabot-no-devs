#include <math.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>

#define PI 3.1415927
#define EPSILON 0.01
#define DIST2(x1, y1, x2, y2)                                                  \
    (((x1) - (x2)) * ((x1) - (x2)) + ((y1) - (y2)) * ((y1) - (y2)))
#define MIN_FP_VALUE 1.0e-100

/* ABI type definitions */

typedef struct {
    unsigned int width;
    unsigned int height;
    unsigned int bpp;
    unsigned char *pixels;
} image_t;


size_t image_size(const image_t *image)
{
    return image->width * image->height * image->bpp;
}

#define NULL_CHECK(value)                                                      \
    if (value == NULL) {                                                       \
        puts("NULL parameter");                                                \
        return;                                                                \
    }

void image_putpixel(const image_t *image, int x, int y, unsigned char r,
                    unsigned char g, unsigned char b, unsigned char a) {
    unsigned char *p;
    if (image == NULL || image->pixels == NULL) {
        return;
    }
    if (x < 0 || y < 0 || x >= (int)image->width || y >= (int)image->height) {
        return;
    }
    p = image->pixels + (x + y * image->width) * image->bpp;
    if (image->bpp == 1) {
        /* convert to grayscale using integer approximation of standard weights */
        /* uses integer arithmetic with coefficients scaled by 256 (77≈0.299×256, 150≈0.587×256, 29≈0.114×256) */
        *p = (unsigned char)((77 * r + 150 * g + 29 * b) >> 8);
    } else {
        *p++ = r;
        *p++ = g;
        *p++ = b;
        if (image->bpp == 4) {
            *p = a;
        }
    }
}


void plasma(const image_t *image, const unsigned char *palette, int x1, int y1, int x2, int y2, int c1, int c2, int c3, int c4, int delta)
{
#define avg(x, y)  (((x)+(y))>>1)
#define step(x, y) (((x)-(y))>>1)
#define bound(x)   if (x>255) x=255; if (x<0) x=0;
#define displac(x, delta)  (x+(rand()%(delta*2)-delta))
    if (x2-x1>1) {
        int dc12=avg(c1, c2);
        int dc13=avg(c1, c3);
        int dc24=avg(c2, c4);
        int dc34=avg(c3, c4);
        int dc  =avg(dc13, dc24);
        int dx=step(x2, x1);
        int dy=step(y2, y1);
        if ((x2-x1>2) && (delta>0))
            dc=displac(dc, delta);
        bound(dc);
        bound(c1);
        bound(c2);
        bound(c3);
        bound(c4);
        delta>>=1;
        plasma(image, palette, x1,    y1,    x1+dx, y1+dy, c1,   dc12, dc13, dc,   delta);
        plasma(image, palette, x1+dx, y1,    x2,    y1+dy, dc12, c2,   dc,   dc24, delta);
        plasma(image, palette, x1,    y1+dy, x1+dx, y2,    dc13, dc,   c3,   dc34, delta);
        plasma(image, palette, x1+dx, y1+dy, x2,    y2,    dc,   dc24, dc34, c4,   delta);
    }
    else {
        unsigned char r, g, b;
        int color_offset = (c1+c2+c3+c4)/4;
        color_offset *= 3;
        unsigned char *pal = (unsigned char *)(palette + color_offset);

        r = *pal++;
        g = *pal++;
        b = *pal;

        image_putpixel(image, x1, y1, r, g, b, 0);
    }
#undef avg
#undef step
#undef bound
}

void render_plasma(const image_t *image, const unsigned char *palette,
                   double zx0_, double zy0_, int maxiter) {

    NULL_CHECK(palette)
    NULL_CHECK(image->pixels)

    plasma(image, palette, 0, 0, image->width-1, image->height-1, 128, 128, 128, 128, maxiter);
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
int bmp_write(unsigned int width, unsigned int height, unsigned char *pixels,
              const char *file_name) {
    unsigned char bmp_header[] = {
        /* BMP header structure: */
        0x42, 0x4d,             /* magic number */
        0x46, 0x00, 0x00, 0x00, /* size of header=70 bytes */
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
    int x, y;

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
        return 1;
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
    }
    fclose(fout);
    return 0;
}

void fill_in_palette(unsigned char *palette) {
    unsigned char *p = palette;
    int i;

    if (palette == NULL) {
        return;
    }

    /* fill in by black color */
    memset(palette, 0, 256 * 3);

    /* green gradient */
    for (i = 0; i < 255; i++) {
        *p++ = i;
        *p++ = i;
        *p++ = i;
    }
}

/*
 * Renders bunch of test images and writes it to both
 * PPM and BMP files.
 */
int render_test_images(void) {
#define WIDTH 512
#define HEIGHT 512

    unsigned char *palette = (unsigned char *)malloc(256 * 3);
    image_t image;

    image.width = WIDTH;
    image.height = HEIGHT;
    image.bpp = 4;
    image.pixels = (unsigned char*)malloc(image_size(&image));

    if (palette == NULL) {
        fprintf(stderr, "Failed to allocate palette memory\n");
        return -1;
    }

    fill_in_palette(palette);

    render_plasma(&image, palette, 0, 0, 100);
    bmp_write(WIDTH, HEIGHT, image.pixels, "test.bmp");

    printf("Main program has ended.\n");

    free(palette);
    return 0;
}

/**
 * Entry point for the program.
 * @returns 0 on successful execution.
 */
#ifndef NO_MAIN
int main(void) {
    int result = render_test_images();

    return result;
}
#endif
