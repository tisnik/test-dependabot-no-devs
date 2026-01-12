#include <stdio.h>
#include <stdlib.h>

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
