#include <stdio.h>
#include <stdlib.h>
#include <string.h>

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

unsigned char true_color_tga_header[] =
{
    0x00,       /* without image ID */
    0x00,       /* color map type: without palette */
    0x02,       /* uncompressed true color image */
    0x00, 0x00, /* start of color palette (it is not used) */
    0x00, 0x00, /* length of color palette (it is not used) */
    0x00,       /* bits per palette entry */
    0x00, 0x00,
    0x00, 0x00, /* image coordinates */
    0x00, 0x00, /* image width */
    0x00, 0x00, /* image height */
    0x18,       /* bits per pixel = 24 */
    0x20        /* picture orientation */
};

int tga_write(unsigned int width, unsigned int height, const unsigned char *pixels,
              const char *file_name)
{
    FILE *fout;
    const unsigned char *p=pixels;
    int i;

    if (pixels==NULL) {
        return 1;
    }

    /* prepare a local header copy to avoid mutating the global template */
    unsigned char header[sizeof true_color_tga_header];
    memcpy(header, true_color_tga_header, sizeof header);

    fout = fopen(file_name, "wb");
    if (!fout)
    {
        return -1;
    }
    /* image size is specified in TGA header */
    header[12]=(width) & 0xff;
    header[13]=(width) >>8;
    header[14]=(height) & 0xff;
    header[15]=(height) >>8;

    /* write TGA header */
    if (fwrite(header, sizeof header, 1, fout) != 1) {
        fclose(fout);
        return -1;
    }

    /* write the whole pixel array into TGA file */
    for (i=0; i<width*height; i++) {
        /* swap RGB to BGR */
        unsigned char bgr[3] = { p[2], p[1], p[0] };
        /* write RGB, but not alpha */
        if (fwrite(bgr, 1, 3, fout) != 3) {
            fclose(fout);
            return -1;
        }
        p+=4; /* skip alpha */
    }

    if (fclose(fout) == EOF)
    {
        return -1;
    }
    return 0;
}

/*
 * Renders bunch of test images and writes it to both
 * PPM and BMP files.
 */
int render_test_images(void) {
#define WIDTH 256
#define HEIGHT 256
    unsigned char *pixels = (unsigned char *)malloc(WIDTH * HEIGHT * 4);
    unsigned char *palette = (unsigned char *)malloc(256*3);

    int i;
    unsigned char *p = palette;
    for (i=0; i<=254; i++) {
        *p++=i*3;
        *p++=i*3;
        *p++=i*3;
    }
    /* last color is black */
    *p++=0;
    *p++=0;
    *p++=0;

    render_test_rgb_image(WIDTH, HEIGHT, pixels, 0);
    tga_write(WIDTH, HEIGHT, pixels, "test_rgb_1.tga");

    return 0;
}

/**
 * Entry point for the program.
 * @returns 0 on successful execution.
 */
int main(void) {
    int result = render_test_images();

    return result;
}
