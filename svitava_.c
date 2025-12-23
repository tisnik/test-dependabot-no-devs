/*

   (C) Copyright 2024, 2025  Pavel Tisnovsky

   All rights reserved. This program and the accompanying materials
   are made available under the terms of the Eclipse Public License v1.0
   which accompanies this distribution, and is available at
   http://www.eclipse.org/legal/epl-v10.html

   Contributors:
       Pavel Tisnovsky

*/

/*
build as shared library: gcc -shared -Wl,-soname,svitava -o svitava.so -fPIC
svitava.c build as executable:
*/

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
#define RGBA 4
#define MAX(a,b) ((a)>(b) ? (a) : (b))
#define MIN(a,b) ((a)<(b) ? (a) : (b))

#define NULL_CHECK(value)                                                      \
    if (value == NULL) {                                                       \
        puts("NULL parameter");                                                \
        return;                                                                \
    }

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

image_t image_create(const unsigned int width, const unsigned int height, const unsigned int bpp)
{
    image_t image;
    image.width = width;
    image.height = height;
    image.bpp = bpp;
    image.pixels = (unsigned char*)malloc(image_size(&image));
    return image;
}

void image_clear(image_t *image)
{
    memset(image->pixels, 0x00, image_size(image));
}

void image_putpixel(image_t *image, int x, int y, unsigned char r, unsigned char g, unsigned char b, unsigned char a)
{
    unsigned char *p;
    if (x<0 || y<0 || x>=(int)image->width || y>=(int)image->height) return;
    p=image->pixels + (x + y * image->width) * image->bpp;
    *p++=r;
    *p++=g;
    *p++=b;
    *p=a;
}

void image_putpixel_max(image_t *image, int x, int y, unsigned char r, unsigned char g, unsigned char b, unsigned char a)
{
    unsigned char *p;
    if (x<0 || y<0 || x>=(int)image->width || y>=(int)image->height) return;
    p=image->pixels + (x + y * image->width) * image->bpp;
    if (*p<r) *p=r;
    p++;
    if (*p<g) *p=g;
    p++;
    if (*p<b) *p=b;
    p++;
    *p=a;
}

void image_getpixel(const image_t *image, int x, int y, unsigned char *r, unsigned char *g, unsigned char *b, unsigned char *a)
{
    unsigned char *p;
    if (x<0 || y<0 || x>=(int)image->width || y>=(int)image->height) return;
    p=image->pixels + (x + y * image->width) * image->bpp;
    *r=*p++;
    *g=*p++;
    *b=*p++;
    *a=*p;
}

void image_hline(image_t *image, int x1, int x2, int y, unsigned char r, unsigned char g, unsigned char b, unsigned char a)
{
    int x, fromX=MIN(x1, x2), toX=MAX(x1, x2);
    for (x=fromX; x<=toX; x++)
    {
        image_putpixel(image, x, y, r, g, b, a);
    }
}

void image_vline(image_t *image, int x, int y1, int y2, unsigned char r, unsigned char g, unsigned char b, unsigned char a)
{
    int y, fromY=MIN(y1, y2), toY=MAX(y1, y2);
    for (y=fromY; y<=toY; y++)
    {
        image_putpixel(image, x, y, r, g, b, a);
    }
}

void image_line(image_t *image, int x1, int y1, int x2, int y2, unsigned char r, unsigned char g, unsigned char b, unsigned char a)
{
    int dx = abs(x2-x1), sx = x1<x2 ? 1 : -1;
    int dy = abs(y2-y1), sy = y1<y2 ? 1 : -1; 
    int err = (dx>dy ? dx : -dy)/2, e2;
    
    while (1){
        image_putpixel(image, x1, y1, r, g, b, a);
        if (x1 == x2 && y1 == y2)
        {
            break;
        }
        e2 = err;
        if (e2 > -dx)
        {
             err -= dy;
             x1 += sx;
        }
        if (e2 < dy)
        {
            err += dx;
            y1 += sy;
        }
    }
}

void image_line_aa(image_t *image, int x1, int y1, int x2, int y2, unsigned char r, unsigned char g, unsigned char b, unsigned char a)
{
    int dx=x2-x1;
    int dy=y2-y1;
    double s, p, e=0.0;
    int x, y, xdelta, ydelta, xpdelta, ypdelta, xp, yp;
    int i, imin, imax;

    if (x1==x2)
    {
        image_vline(image, x1, y1, y2, r, g, b, a);
        return;
    }

    if (y1==y2)
    {
        image_hline(image, x1, x2, y1, r, g, b, a);
        return;
    }

    if (x1>x2)
    {
        x1=x1^x2; x2=x1^x2; x1=x1^x2;
        y1=y1^y2; y2=y1^y2; y1=y1^y2;
    }

    if (abs(dx)>abs(dy))
    {
        s=(double)dy/(double)dx;
        imin=x1;  imax=x2;
        x=x1;     y=y1;
        xdelta=1; ydelta=0;
        xpdelta=0;
        xp=0;
        if (y2>y1)
        {
            ypdelta=1;
            yp=1;
        }
        else
        {
            s=-s;
            ypdelta=-1;
            yp=-1;
        }
    }
    else
    {
        s=(double)dx/(double)dy;
        xdelta=0; ydelta=1;
        ypdelta=0;
        yp=0;
        if (y2>y1)
        {
            imin=y1;    imax=y2;
            x=x1;       y=y1;
            xpdelta=1;
            xp=1;
        }
        else
        {
            s=-s;
            imin=y2;    imax=y1;
            x=x2;       y=y2;
            xpdelta=-1;
            xp=-1;
        }
    }
    p=s*256.0;
    for (i=imin; i<=imax; i++) {
        int c1=(int)e;
        int c2=255-c1;
        image_putpixel_max(image, x+xp, y+yp, (r*c1)/255, (g*c1)/255, (b*c1)/255, a);
        image_putpixel_max(image, x, y, (r*c2)/255, (g*c2)/255, (b*c2)/255, a);
        e=e+p;
        x+=xdelta;
        y+=ydelta;
        if (e>=256.0) {
            e-=256.0;
            x+=xpdelta;
            y+=ypdelta;
        }
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
        0x46, 0x00, 0x00, 0x00, /* size in bytes (placeholder, unused) */
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

int test_drawing_operations(void)
{
#define WIDTH 512
#define HEIGHT 512

    image_t image1 = image_create(512, 512, RGBA);

    image_clear(&image1);
    {
        int x, y;
        for (y=0; y<2; y++) {
            for (x=0; x<2; x++) {
                image_putpixel(&image1, 20+x, 20+y, 255, 0, 0, 0);
                image_putpixel(&image1, 40+x, 20+y, 0, 255, 0, 0);
                image_putpixel(&image1, 20+x, 40+y, 0, 0, 255, 0);
                image_putpixel(&image1, 40+x, 40+y, 255, 255, 255, 0);
            }
        }
    }

    image_hline(&image1, 10, 500, 100, 255, 100, 100, 0);
    image_vline(&image1, 10, 110, 500, 100, 100, 255, 0);
    {
        int y;
        for (y=120; y<300; y+=20) {
            image_line(&image1, 20, 120, 500, y, 255, 255, 255, 0);
        }
    }
    {
        int y;
        for (y=320; y<500; y+=20) {
            image_line_aa(&image1, 20, 320, 500, y, 255, 255, 255, 0);
        }
    }

    bmp_write(WIDTH, HEIGHT, image1.pixels, "image1.bmp");
    free(image1.pixels);  /* Also: memory leak without this */
    return 0;
}


/**
 * Entry point for the program.
 * @returns 0 on successful execution.
 */
#ifndef NO_MAIN
int main(void) {
    /*
    int result = render_test_images();
    int result = render_blended_images();
    */
    int result = test_drawing_operations();

    return result;
}
#endif
