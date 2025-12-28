void composite_horizontal_interlace(const image_t *src1, const image_t *src2, image_t *dest) {
    unsigned int i, j;
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

void composite_vertical_interlace(const image_t *src1, const image_t *src2, image_t *dest) {
    unsigned int i, j;
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

void composite_interlace(const image_t *src1, const image_t *src2, image_t *dest) {
    unsigned int i, j;
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

void composite_blend(const image_t *src1, const image_t *src2, image_t *dest) {
    unsigned int i, j;
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

int image_export_tga(unsigned int width, unsigned int height,
              const unsigned char *pixels, const char *file_name) {
    FILE                *fout;
    const unsigned char *p = pixels;
    unsigned char        header[sizeof true_color_tga_header];
    int                  i;

    if (pixels == NULL) {
        return 1;
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

void palette_greens(unsigned char *palette) {
    unsigned char *p = palette;
    int            i;

    if (palette == NULL) {
        return;
    }

    /* fill in by black color */
    memset(palette, 0, 256 * 3);

    /* green gradient */
    for (i = 0; i < 32; i++) {
        *p++ = 0;
        *p++ = 4 + i * 6;
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
        int c = 252 - i * 6;
        if (c < 0) {
            c = 0;
        }
        *p++ = 252;
        *p++ = 252;
        *p++ = c;
    }

    /* gradient from yellow to green */
    for (i = 0; i < 48; i++) {
        int c = 252 - i * 6;
        if (c < 0) {
            c = 0;
        }
        *p++ = c;
        *p++ = 252;
        *p++ = 0;
    }

    /* gradient green to black */
    for (i = 0; i < 48; i++) {
        int c = 252 - i * 6;
        if (c < 0) {
            c = 0;
        }
        *p++ = 0;
        *p++ = c;
        *p++ = 0;
    }
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
    /*
    int result = render_test_images();
    int result = render_blended_images();
    */
    int result = test_drawing_operations();

    return result;
}
#endif
