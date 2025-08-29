#include <stdio.h>
#include <stdlib.h>
#include <string.h>

typedef struct {
    unsigned int width;
    unsigned int height;
    unsigned char *pixels;
} image_t;

#define NULL_CHECK(value)                                                      \
    if (value == NULL) {                                                       \
        fprintf(stderr, "NULL parameter: %s\n", #value);                       \
        return;                                                                \
    }

/**
 * Writes an RGB color from the palette at the specified index into the pixel
 * buffer and advances the pixel pointer by 4 bytes.
 */
void putpixel(unsigned char **pixel, const unsigned char *palette,
        int color_index) {
    int color_offset;
    unsigned char *pal;

    if (color_index < 0) {
        color_index = 0;
    }
    if (color_index > 255) {
        color_index = 255;
    }

    color_offset = color_index * 3;
    pal = (unsigned char *)(palette + color_offset);

    *(*pixel)++ = *pal++;
    *(*pixel)++ = *pal++;
    *(*pixel)++ = *pal;
    (*pixel)++;
}

/**
 * Renders the Julia set fractal into a pixel buffer using a specified
 * color palette.
 *
 * Iterates over each pixel, maps it to a point in the complex plane, and
 * computes the escape time for the Julia set. The iteration count
 * determines the color index used from the palette.
 */
void render_julia(const image_t *image, const unsigned char *palette,
                  const double cx, const double cy, unsigned int maxiter) {
    int x, y;
    double zx0, zy0;
    double xmin = -1.5, ymin = -1.5, xmax = 1.5, ymax = 1.5;
    unsigned char *p;

    NULL_CHECK(image)
    NULL_CHECK(palette)
    NULL_CHECK(image->pixels)

    p = image->pixels;

    zy0 = ymin;
    for (y = 0; y < image->height; y++) {
        zx0 = xmin;
        for (x = 0; x < image->width; x++) {
            double zx = zx0;
            double zy = zy0;
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
            putpixel(&p, palette, i);
            zx0 += (xmax - xmin) / image->width;
        }
        zy0 += (ymax - ymin) / image->height;
    }
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
    if (fwrite(bmp_header, sizeof(bmp_header), 1, fout) != 1) {
        fclose(fout);
        return 1;
    }

    /* write the whole pixel array into BMP file */
    for (y = height - 1; y >= 0; y--) {
        /* pointer to the 1st pixel on scan line */
        unsigned char *p = pixels + y * width * 4;
        for (x = 0; x < width; x++) {
            /* swap RGB color components as required by file format */
            if (fwrite(p + 2, 1, 1, fout) != 1 ||
                fwrite(p + 1, 1, 1, fout) != 1 ||
                fwrite(p, 1, 1, fout) != 1) {
                fclose(fout);
                return 1;
            }
            /* move to next pixel on scan line */
            p += 4;
        }
    }
    if (fclose(fout) == EOF) {
        return 1;
    }
    return 0;
}

unsigned char *generate_palette(void) {
    unsigned char *palette = (unsigned char *)malloc(256 * 3);
    unsigned char *p = palette;
    int i;

    if (palette == NULL) {
        return NULL;
    }

    /* fill in by black color */
    memset(palette, 0, 256 * 3);

    /* green gradient */
    for (i = 0; i < 32; i++) {
        *p++ = 0;
        *p++ = 4 + i*6;
        *p++ = 0;
    }

    /* gradient from green to yellow */
    for (i = 0; i < 32; i++) {
        *p++ = 4 + i * 6;
        *p++ = i * 2 < 52 ? 200 + i * 2 : 252;
        *p++ = 0;
    }

    /* gradient from yellow to white */
    for (i = 0; i < 32; i++) {
        *p++ = i * 2 < 52 ? 200 + i * 2 : 252;
        *p++ = 252;
        *p++ = i * 6;
    }

    /* gradient from white to yellow */
    for (i = 0; i < 48; i++) {
        *p++ = 252;
        *p++ = 252;
        *p++ = 252 - i * 6;
    }
    
    /* gradient from yellow to green */
    for (i = 0; i < 48; i++) {
        *p++ = 252 - i * 6;
        *p++ = 252;
        *p++ = 0;
    }
    
    /* gradient green to black */
    for (i = 0; i < 48; i++) {
        *p++ = 0;
        *p++ = 252 - i * 6;
        *p++ = 0;
    }

    return palette;
}

/*
 * Renders test image and writes it to BMP file.
 */
int render_test_image(void) {
#define WIDTH 512
#define HEIGHT 512
    unsigned char *pixels = (unsigned char *)malloc(WIDTH * HEIGHT * 4);
    unsigned char *palette = generate_palette();
    image_t image;
    int result;

    if (!pixels) {
        return -1;
    }

    if (!palette) {
        free(pixels);
        return -1;
    }

    image.width = WIDTH;
    image.height = HEIGHT;
    image.pixels = pixels;


    render_julia(&image, palette, -0.207190825, 0.676656625, 255);
    result = bmp_write(WIDTH, HEIGHT, pixels, "julia.bmp");

    free(pixels);
    free(palette);
    return result;
}

/**
 * Entry point for the program.
 * @returns 0 on successful execution.
 */
int main(void) {
    int result = render_test_image();

    return result;
}
