/*
build as shared library: gcc -shared -Wl,-soname,testlib -o testlib.so -fPIC
testlib.c build as executable:
*/

#include <stdio.h>
#include <stdlib.h>

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

int main(void) {
#define WIDTH 256
#define HEIGHT 256
    unsigned char *pixels = (unsigned char *)malloc(WIDTH * HEIGHT * 4);
    render_test_rgb_image(WIDTH, HEIGHT, pixels, 0);
    ppm_write_ascii(WIDTH, HEIGHT, pixels, "test_rgb_1.ppm");
    return 0;
}
