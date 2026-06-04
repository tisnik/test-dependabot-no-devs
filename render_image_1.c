/*
build as shared library: gcc -shared -Wl,-soname,testlib -o testlib.so -fPIC
testlib.c build as executable:
*/

#include <stdio.h>
#include <stdlib.h>

/**
 * Fills a pixel buffer with a test RGBA image pattern.
 *
 * Sets each pixel's red channel to its x-coordinate, green channel to the specified value, blue channel to its y-coordinate, and leaves the alpha channel unused.
 * The buffer must be preallocated with at least width * height * 4 bytes.
 * @param width Image width in pixels.
 * @param height Image height in pixels.
 * @param pixels Pointer to the RGBA pixel buffer to fill.
 * @param green Value to assign to the green channel for all pixels.
 */
void render_test_rgb_image(unsigned int width, unsigned int height,
                           unsigned char *pixels, unsigned char green) {
    unsigned int i, j;
    unsigned char *p = pixels;

    for (j = 0; j < height; j++) {
        for (i = 0; i < width; i++) {
            *p++ = i;
            *p++ = green;
            *p++ = j;
            p++;
        }
    }
}

/**
 * Writes pixel data to a file stream in ASCII PPM (P3) format.
 *
 * Outputs the image header and RGB values for each pixel, reading from the provided buffer in bottom-to-top row order. The alpha channel in the buffer is ignored.
 */
void ppm_write_ascii_to_stream(unsigned int width, unsigned int height,
                               unsigned char *pixels, FILE *fout) {
    int x, y;
    unsigned char r, g, b;
    unsigned char *p = pixels;

    /* header */
    fprintf(fout, "P3 %d %d 255\n", width, height);

    /* pixel array */
    for (y = height - 1; y >= 0; y--) {
        for (x = 0; x < width; x++) {
            r = *p++;
            g = *p++;
            b = *p++;
            p++;
            fprintf(fout, "%d %d %d\n", r, g, b);
        }
    }
}

/**
 * Writes pixel data to a file in ASCII PPM (P3) format.
 *
 * Opens the specified file for writing, writes the image data in ASCII PPM format using the provided pixel buffer, and closes the file.
 * Returns 0 on success, or -1 if the file cannot be opened or closed.
 * @param width Image width in pixels.
 * @param height Image height in pixels.
 * @param pixels Pointer to the RGBA pixel buffer (only RGB channels are written).
 * @param file_name Name of the output file.
 * @return 0 on success, -1 on failure.
 */
int ppm_write_ascii(unsigned int width, unsigned int height,
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
 * Generates a 256x256 test RGB image and writes it to "test_rgb_1.ppm" in ASCII PPM format.
 * @returns 0 on successful completion.
 */
int main(void) {
#define WIDTH 256
#define HEIGHT 256
    unsigned char *pixels = (unsigned char *)malloc(WIDTH * HEIGHT * 4);
    render_test_rgb_image(WIDTH, HEIGHT, pixels, 0);
    ppm_write_ascii(WIDTH, HEIGHT, pixels, "test_rgb_1.ppm");
    return 0;
}
