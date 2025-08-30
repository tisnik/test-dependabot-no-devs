/*
  SDL Fractal Viewer

   Copyright (C) 2019  Pavel Tisnovsky

SDL Fractal Viewer is free software; you can redistribute it
and/or modify it under the terms of the GNU General Public License as
published by the Free Software Foundation; either version 2, or (at your
option) any later version.

SDL Fractal Viewer is distributed in the hope that it will be
useful, but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
General Public License for more details.

You should have received a copy of the GNU General Public License along
with SDL Fractal Viewer; see the file COPYING.  If not, write
to the Free Software Foundation, Inc., 51 Franklin Street, Fifth Floor,
Boston, MA 02110-1301 USA.

Linking this library statically or dynamically with other modules is
making a combined work based on this library.  Thus, the terms and
conditions of the GNU General Public License cover the whole
combination.

As a special exception, the copyright holders of this library give you
permission to link this library with independent modules to produce an
executable, regardless of the license terms of these independent
modules, and to copy and distribute the resulting executable under
terms of your choice, provided that you also meet, for each linked
independent module, the terms and conditions of the license of that
module.  An independent module is a module which is not derived from
or based on this library.  If you modify this library, you may extend
this exception to your version of the library, but you are not
obligated to do so.  If you do not wish to do so, delete this
exception statement from your version.
*/

#include <math.h>

#include <SDL/SDL.h>

#include "gfx.h"

#define ABS(a) ((a)<0 ? -(a) : (a) )
#define MAX(a,b) ((a)>(b) ? (a) : (b))
#define MIN(a,b) ((a)<(b) ? (a) : (b))

static SDL_Surface *screen_surface = NULL;
static SDL_Surface *bitmap_font_surface = NULL;

/*
TTF_Font *font24pt = NULL;
TTF_Font *font36pt = NULL;
TTF_Font *font48pt = NULL;
*/

/**
 * Initialize the SDL video subsystem and create the module's main screen surface.
 *
 * Initializes SDL's video subsystem and attempts to set the global screen surface (screen_surface)
 * using SDL_SetVideoMode. On success returns 0; on failure prints an error to stderr and returns 1.
 * Note: the current implementation ignores the width, height, and bpp arguments and always requests
 * a 640x480, 32 bpp surface with hardware surface and double buffering. If fullscreen is non-zero
 * the SDL_FULLSCREEN flag is also requested.
 *
 * @param fullscreen If non-zero, request fullscreen mode when setting the video mode.
 * @param width Ignored by this implementation (present for API compatibility).
 * @param height Ignored by this implementation (present for API compatibility).
 * @param bpp Ignored by this implementation (present for API compatibility).
 * @return 0 on success, 1 on failure.
 */
int gfx_initialize(int fullscreen, int width, int height, int bpp)
{
    if (SDL_Init(SDL_INIT_VIDEO) < 0)
    {
        fprintf(stderr, "Error initializing SDL: %s\n", SDL_GetError());
        return 1;
    }
    if (fullscreen)
    {
        screen_surface = SDL_SetVideoMode(640, 480, 32, SDL_HWSURFACE | SDL_DOUBLEBUF | SDL_FULLSCREEN | SDL_ANYFORMAT);
    }
    else
    {
        screen_surface = SDL_SetVideoMode(640, 480, 32, SDL_HWSURFACE | SDL_DOUBLEBUF | SDL_ANYFORMAT);
    }
    if (!screen_surface)
    {
        fprintf(stderr, "Error setting video mode: %s\n", SDL_GetError());
        return 1;
    }
    /*
    if (TTF_Init() < 0)
    {
        gfx_finalize();
        fprintf(stderr, "Error initializing TTF moduler");
        return 1;
    }*/
    /*
    font24pt = TTF_OpenFont("data/handwritten.ttf", 24);
    font36pt = TTF_OpenFont("data/handwritten.ttf", 36);
    font48pt = TTF_OpenFont("data/handwritten.ttf", 48);
    */
    return 0;
}

/**
 * Free the module's primary screen surface.
 *
 * Releases the internally stored SDL_Surface pointed to by the module's
 * screen_surface (calls SDL_FreeSurface). If no screen surface is set,
 * the function has no effect.
 */
void gfx_finalize(void)
{
    SDL_FreeSurface(screen_surface);
}

/**
 * Set the module-wide primary screen surface.
 *
 * The provided SDL_Surface becomes the surface used by screen-specific
 * drawing helpers (e.g., gfx_putpixel_screen, gfx_flip). This function
 * does not allocate or free the surface; ownership and lifetime remain
 * the caller's responsibility.
 *
 * @param screen SDL_Surface to use as the primary screen surface.
 */
void gfx_set_screen_surface(SDL_Surface *screen)
{
    screen_surface = screen;
}

/**
 * Get the current primary screen surface used by the gfx module.
 *
 * Returns a pointer to the module-owned SDL_Surface that represents the main
 * display surface (set via gfx_initialize or gfx_set_screen_surface). The
 * returned surface is owned by the gfx module and must not be freed by callers.
 *
 * @return Pointer to the current screen SDL_Surface, or NULL if no surface is set.
 */
SDL_Surface* gfx_get_screen_surface(void)
{
    return screen_surface;
}

/**
 * Create a new 32-bit RGBA SDL surface.
 *
 * Allocates and returns an SDL surface with the given width and height using
 * 32 bits per pixel and the RGB byte masks (R: 0x00ff0000, G: 0x0000ff00,
 * B: 0x000000ff). The alpha mask is set to 0 (no alpha channel).
 *
 * @param width Width of the new surface in pixels.
 * @param height Height of the new surface in pixels.
 * @returns Pointer to the newly created SDL_Surface on success, or NULL on failure.
 */
SDL_Surface* gfx_create_surface(int width, int height)
{
    return SDL_CreateRGBSurface(SDL_SWSURFACE, width, height, 32, 0x00ff0000, 0x0000ff00, 0x000000ff, 0x00000000);
}

/**
 * Blit a surface onto the module's screen surface at the given coordinates.
 *
 * The entire source surface is copied (source rect = NULL) and placed with its
 * top-left corner at (x, y) on the internal screen surface.
 *
 * @param surface Source SDL_Surface to blit.
 * @param x Horizontal destination coordinate (pixels).
 * @param y Vertical destination coordinate (pixels).
 */
void gfx_bitblt(SDL_Surface *surface, const int x, const int y)
{
    SDL_Rect dst_rect;
    dst_rect.x = x;
    dst_rect.y = y;
    SDL_BlitSurface(surface, NULL, screen_surface, &dst_rect);
}

/**
 * Flip the module screen surface to update the display.
 *
 * Calls SDL_Flip on the internal screen surface previously set by gfx_initialize
 * or gfx_set_screen_surface. The screen surface must be initialized (non-NULL)
 * before calling.
 */
void gfx_flip()
{
    SDL_Flip(screen_surface);
}

/**
 * Fill the entire configured screen surface with a color.
 *
 * Fills the module's global screen_surface with the given 32-bit pixel value.
 * The color must be provided in the surface's pixel format (for example, produced
 * by SDL_MapRGB/SDL_MapRGBA). If the internal screen surface is NULL, the call
 * has no effect.
 *
 * @param color 32-bit pixel value to use for the fill.
 */
void gfx_clear_screen(Uint32 color)
{
    SDL_FillRect(screen_surface, NULL, color);
}

/**
 * Set a pixel on an SDL_Surface at (x,y) to the specified RGB color.
 *
 * If (x,y) is outside the surface bounds the function does nothing.
 * Only 24-bit and 32-bit surfaces are supported; for 32-bit surfaces the
 * alpha channel is ignored.
 *
 * @param surface Target SDL_Surface.
 * @param x Horizontal pixel coordinate.
 * @param y Vertical pixel coordinate.
 * @param r Red component (0-255).
 * @param g Green component (0-255).
 * @param b Blue component (0-255).
 */
void gfx_putpixel(SDL_Surface *surface, int x, int y, unsigned char r, unsigned char g, unsigned char b)
{
    if (x>=0 && x< surface->w && y>=0 && y < surface->h)
    {
        if (surface->format->BitsPerPixel == 24)
        {
            Uint8 *pixel = (Uint8 *)surface->pixels;
            pixel += x*3;
            pixel += y*surface->pitch;
            *pixel++ = b;
            *pixel++ = g;
            *pixel   = r;
        }
        if (surface->format->BitsPerPixel == 32)
        {
            Uint8 *pixel = (Uint8 *)surface->pixels;
            pixel += x*4;
            pixel += y*surface->pitch;
            *pixel++ = b;
            *pixel++ = g;
            *pixel   = r;
        }
    }
}

/**
 * Plot a single pixel on the module's screen surface.
 *
 * Draws a pixel at the given (x, y) coordinates on the global screen surface
 * using the provided RGB components. Coordinates outside the surface bounds
 * are ignored; color components are in the 0–255 range.
 *
 * @param x X coordinate in pixels.
 * @param y Y coordinate in pixels.
 * @param r Red component (0-255).
 * @param g Green component (0-255).
 * @param b Blue component (0-255).
 */
void gfx_putpixel_screen(int x, int y, unsigned char r, unsigned char g, unsigned char b)
{
    gfx_putpixel(screen_surface, x, y, r, g, b);
}

/**
 * Draw a horizontal line on the given surface.
 *
 * Draws a 1-pixel-high horizontal line from x1 to x2 at vertical position y.
 * The endpoints are inclusive; x1 and x2 may be provided in any order.
 *
 * @param surface Target SDL_Surface to draw on.
 * @param x1 One horizontal endpoint (order-independent).
 * @param x2 The other horizontal endpoint (order-independent).
 * @param y Vertical coordinate of the line.
 * @param r Red component (0–255).
 * @param g Green component (0–255).
 * @param b Blue component (0–255).
 */
void gfx_hline(SDL_Surface *surface, int x1, int x2, int y, unsigned char r, unsigned char g, unsigned char b)
{
    int x;
    int fromX = MIN(x1, x2);
    int toX = MAX(x1, x2);
    for (x = fromX; x <= toX; x++)
    {
        gfx_putpixel(surface, x, y, r, g, b);
    }
}

/**
 * Draw a vertical line on a surface.
 *
 * Draws an inclusive vertical line of pixels from y1 to y2 at column x on the
 * given SDL_Surface using the specified RGB color. The order of y1 and y2 is
 * irrelevant (the function will draw between the min and max). Pixels outside
 * the surface bounds are skipped (relies on gfx_putpixel's bounds checks).
 *
 * @param surface Target SDL surface to draw into.
 * @param x X coordinate (column) where the line will be drawn.
 * @param y1 One end Y coordinate of the line (inclusive).
 * @param y2 Other end Y coordinate of the line (inclusive).
 * @param r Red component (0-255).
 * @param g Green component (0-255).
 * @param b Blue component (0-255).
 */
void gfx_vline(SDL_Surface *surface, int x,  int y1, int y2, unsigned char r, unsigned char g, unsigned char b)
{
    int y;
    int fromY = MIN(y1, y2);
    int toY = MAX(y1, y2);
    for (y = fromY; y <= toY; y++)
    {
        gfx_putpixel(surface, x, y, r, g, b);
    }
}

/**
 * Draw a straight RGB line between two points on the given surface.
 *
 * Draws a continuous line from (x1, y1) to (x2, y2) on `surface`, setting each pixel's
 * color to the provided red/green/blue components. The endpoints are inclusive.
 * Uses an integer line rasterization algorithm (Bresenham-style) and works for all slopes.
 *
 * @param surface Target SDL_Surface to draw onto.
 * @param x1 X coordinate of the start point.
 * @param y1 Y coordinate of the start point.
 * @param x2 X coordinate of the end point.
 * @param y2 Y coordinate of the end point.
 * @param r Red component (0–255).
 * @param g Green component (0–255).
 * @param b Blue component (0–255).
 */
void gfx_line(SDL_Surface *surface, int x1, int y1, int x2, int y2, unsigned char r, unsigned char g, unsigned char b)
{
    int dx = abs(x2-x1), sx = x1<x2 ? 1 : -1;
    int dy = abs(y2-y1), sy = y1<y2 ? 1 : -1; 
    int err = (dx>dy ? dx : -dy)/2, e2;
    
    while (1){
        gfx_putpixel(surface, x1, y1, r, g, b);
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

/**
 * Draw an anti-aliased line between two points onto an SDL surface.
 *
 * Uses a simple Xiaolin Wu–style approach (per-axis major/minor stepping) to
 * blend two neighboring pixels per step and reduce aliasing. If the line is
 * perfectly vertical or horizontal, the function delegates to gfx_vline or
 * gfx_hline respectively. If x1 > x2 the endpoints are swapped to iterate
 * left-to-right. Pixel writes are performed via gfx_putpixel (which performs
 * its own bounds checks).
 *
 * @param surface Target SDL_Surface to draw on.
 * @param x1 X coordinate of the first endpoint.
 * @param y1 Y coordinate of the first endpoint.
 * @param x2 X coordinate of the second endpoint.
 * @param y2 Y coordinate of the second endpoint.
 * @param r Red intensity (0–255).
 * @param g Green intensity (0–255).
 * @param b Blue intensity (0–255).
 *
 * Notes:
 * - Color components equal to 255 are treated as fully opaque for both the
 *   main and adjacent pixel (i.e., no attenuation is applied when a component
 *   is 255).
 * - The function does not perform additional clipping beyond what gfx_putpixel
 *   enforces.
 */
void gfx_aa_line(SDL_Surface *surface, int x1, int y1, int x2, int y2, unsigned char r, unsigned char g, unsigned char b)
{
    int dx=x2-x1;
    int dy=y2-y1;
    double s, p, e=255.0;
    int x, y, xdelta, ydelta, xpdelta, ypdelta, xp, yp;
    int i, imin, imax;

    if (x1==x2)
    {
        gfx_vline(surface, x1, y1, y2, r, g, b);
        return;
    }

    if (y1==y2)
    {
        gfx_hline(surface, x1, x2, y1, r, g, b);
        return;
    }

    if (x1>x2)
    {
        x1=x1^x2; x2=x1^x2; x1=x1^x2;
        y1=y1^y2; y2=y1^y2; y1=y1^y2;
    }

    if (ABS(dx)>ABS(dy))
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
        int c1 = (int)e;
        int c2 = 255-c1;
        int r1 = r == 255 ? 255 : c1;
        int r2 = r == 255 ? 255 : c2;
        int g1 = g == 255 ? 255 : c1;
        int g2 = g == 255 ? 255 : c2;
        int b1 = b == 255 ? 255 : c1;
        int b2 = b == 255 ? 255 : c2;
        gfx_putpixel(surface, x+xp, y+yp, r1, g1, b1);
        gfx_putpixel(surface, x, y, r2, g2, b2);
        e=e-p;
        x+=xdelta;
        y+=ydelta;
        if (e<0.0) {
            e+=256.0;
            x+=xpdelta;
            y+=ypdelta;
        }
    }
}

/**
 * Draw a horizontal line on the module's current screen surface.
 *
 * Draws a horizontal line between x1 and x2 (inclusive) at vertical coordinate y
 * onto the global screen surface set by gfx_set_screen_surface(). The function
 * accepts RGB color components in the 0–255 range. The order of x1 and x2 is
 * not significant (the implementation will draw between the min and max).
 *
 * @param x1   X coordinate of the first endpoint (inclusive).
 * @param x2   X coordinate of the second endpoint (inclusive).
 * @param y    Y coordinate at which to draw the line.
 * @param r    Red component (0–255).
 * @param g    Green component (0–255).
 * @param b    Blue component (0–255).
 */
void gfx_hline_screen(int x1, int x2, int y, unsigned char r, unsigned char g, unsigned char b)
{
    gfx_hline(screen_surface, x1, x2, y, r, g, b);
}

/**
 * Draw a vertical line on the current screen surface.
 *
 * Draws a vertical line at column `x` from `y1` to `y2` (inclusive) on the module's
 * screen surface using the specified RGB color. Coordinates outside the surface
 * bounds are ignored (drawing is clipped).
 *
 * @param x   X coordinate (column) of the line.
 * @param y1  Starting Y coordinate.
 * @param y2  Ending Y coordinate.
 * @param r   Red component (0-255).
 * @param g   Green component (0-255).
 * @param b   Blue component (0-255).
 */
void gfx_vline_screen(int x, int y1, int y2, unsigned char r, unsigned char g, unsigned char b)
{
    gfx_vline(screen_surface, x, y1, y2, r, g, b);
}

/**
 * Draw a straight (Bresenham) line on the current screen surface.
 *
 * Wrapper around gfx_line that draws a line between (x1, y1) and (x2, y2)
 * on the module's configured screen surface using the specified RGB color.
 */
void gfx_line_screen(int x1, int y1, int x2, int y2, unsigned char r, unsigned char g, unsigned char b)
{
    gfx_line(screen_surface, x1, y1, x2, y2, r, g, b);
}

/**
 * Draw an anti-aliased line on the current screen surface.
 *
 * Renders an anti-aliased line between (x1, y1) and (x2, y2) using the
 * globally configured screen surface.
 *
 * @param x1 X coordinate of the line start (pixels).
 * @param y1 Y coordinate of the line start (pixels).
 * @param x2 X coordinate of the line end (pixels).
 * @param y2 Y coordinate of the line end (pixels).
 * @param r  Red component of the line color (0-255).
 * @param g  Green component of the line color (0-255).
 * @param b  Blue component of the line color (0-255).
 */
void gfx_aa_line_screen(int x1, int y1, int x2, int y2, unsigned char r, unsigned char g, unsigned char b)
{
    gfx_aa_line(screen_surface, x1, y1, x2, y2, r, g, b);
}

/**
 * Set the bitmap font surface used for bitmap-font text rendering.
 *
 * The provided SDL_Surface should contain the glyph atlas used by
 * gfx_print_char_bitmap_font / gfx_print_string_bitmap_font. The function
 * stores the pointer internally; it does not take ownership or perform a
 * copy. The caller is responsible for ensuring the surface remains valid
 * for as long as the gfx module may use it.
 *
 * @param surface SDL_Surface containing the bitmap font glyphs (or NULL to unset).
 */
void gfx_set_bitmap_font_surface(SDL_Surface *surface)
{
    bitmap_font_surface = surface;
}

/**
 * Render a single character from the bitmap font onto a surface.
 *
 * Renders the glyph corresponding to the ASCII character `ch` from the module-level
 * bitmap_font_surface into `surface` at position (x, y). The function maps the
 * ASCII code to a glyph index by subtracting 32; characters with resulting index
 * <= 0 are ignored (no drawing). The glyph rectangle uses BITMAP_FONT_CHARACTER_WIDTH
 * and BITMAP_FONT_CHARACTER_HEIGHT and is blitted with SDL_BlitSurface.
 *
 * @param surface Target SDL_Surface to draw the glyph onto.
 * @param x X coordinate (left) on the target surface where the glyph will be placed.
 * @param y Y coordinate (top) on the target surface where the glyph will be placed.
 * @param ch ASCII character to draw (expects printable characters; index = ch - 32).
 */
void gfx_print_char_bitmap_font(SDL_Surface *surface, int x, int y, char ch)
{
    SDL_Rect src_rect;
    SDL_Rect dst_rect;
    ch-=32;
    if (ch <= 0)
    {
        return;
    }
    dst_rect.x = x;
    dst_rect.y = y;
    src_rect.x = 0;
    src_rect.y = BITMAP_FONT_CHARACTER_HEIGHT * ch;
    src_rect.w = BITMAP_FONT_CHARACTER_WIDTH;
    src_rect.h = BITMAP_FONT_CHARACTER_HEIGHT;
    SDL_BlitSurface(bitmap_font_surface, &src_rect, surface, &dst_rect);
}

/**
 * Render a null-terminated string to a surface using the module's bitmap font.
 *
 * Renders each character in `str` by calling gfx_print_char_bitmap_font and advances
 * the drawing X position by BITMAP_FONT_CHARACTER_WIDTH for each glyph.
 * The bitmap font surface must have been set with gfx_set_bitmap_font_surface()
 * before calling this function.
 *
 * @param surface Destination SDL_Surface to draw the string onto.
 * @param x X coordinate of the leftmost character (in pixels).
 * @param y Y coordinate of the top of the characters (in pixels).
 * @param str Null-terminated C string to render.
 */
void gfx_print_string_bitmap_font(SDL_Surface *surface, int x, int y, char *str)
{
    char *p;
    for (p = str; *p; p++)
    {
        gfx_print_char_bitmap_font(surface, x, y, *p);
        x += BITMAP_FONT_CHARACTER_WIDTH;
    }
}

/**
 * Render a null-terminated string onto the module's primary screen surface using the bitmap font.
 *
 * The text is drawn starting at pixel coordinates (x, y). The bitmap font surface must be set
 * beforehand via gfx_set_bitmap_font_surface(). This is a thin wrapper that targets the
 * internal screen surface.
 * @param x X coordinate in pixels where rendering starts.
 * @param y Y coordinate in pixels where rendering starts.
 * @param str Null-terminated C string to render.
 */
void gfx_print_string_bitmap_font_screen(int x, int y, char *str)
{
    gfx_print_string_bitmap_font(screen_surface, x, y, str);
}

/*
 *
 */
/*
void gfx_print_string_ttf(SDL_Surface *surface, TTF_Font *font, int x, int y, char *str)
{
    SDL_Rect dst_rect;
    SDL_Color color;
    SDL_Surface *font_surface;
    dst_rect.x = x;
    dst_rect.y = y;
    color.r = 0x00;
    color.g = 0x00;
    color.b = 0x00;
    font_surface = TTF_RenderText_Blended(font, str, color);
    SDL_BlitSurface(font_surface, NULL, surface, &dst_rect);
    SDL_FreeSurface(font_surface);
}*/

/*
 *
 */
/*
void gfx_print_string_ttf_screen(TTF_Font *font, int x, int y, char *str){
    gfx_print_string_ttf(screen_surface, font, x, y, str);
}*/

