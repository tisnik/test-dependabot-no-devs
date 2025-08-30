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

#include <SDL2/SDL.h>

#include "gfx.h"

#define ABS(a) ((a)<0 ? -(a) : (a) )
#define MAX(a,b) ((a)>(b) ? (a) : (b))
#define MIN(a,b) ((a)<(b) ? (a) : (b))

static SDL_Surface *screen_surface = NULL;
static SDL_Surface *bitmap_font_surface = NULL;
static SDL_Window *window = NULL;

/*
TTF_Font *font24pt = NULL;
TTF_Font *font36pt = NULL;
TTF_Font *font48pt = NULL;
*/

/**
 * Initialize SDL video, create an application window, and obtain the window surface.
 *
 * Attempts to initialize the SDL video subsystem, create a window titled "Example"
 * with the given width/height, and store the resulting SDL_Surface in the module's
 * global screen_surface. On success the function returns 0 and the global `window`
 * and `screen_surface` are set; on failure it prints an error to stderr and returns 1.
 *
 * @param fullscreen If non-zero, intended to request fullscreen mode (currently ignored).
 * @param width Width in pixels for the created window.
 * @param height Height in pixels for the created window.
 * @param bpp Bits-per-pixel request (present for API compatibility; currently unused).
 * @return 0 on success, 1 on failure.
 */
int gfx_initialize(int fullscreen, int width, int height, int bpp)
{
    window = NULL;
    if (SDL_Init(SDL_INIT_VIDEO) < 0) {
        fprintf(stderr, "Error initializing SDL: %s\n", SDL_GetError());
        return 1;
    } else {
        puts("SDL_Init ok");
    }

    window = SDL_CreateWindow( "Example", SDL_WINDOWPOS_UNDEFINED, SDL_WINDOWPOS_UNDEFINED, width, height, SDL_WINDOW_SHOWN );
    if (!window) {
        puts("Error creating window");
        puts(SDL_GetError());
        return 1;
    } else {
        puts("SDL_CreateWindow ok");
    }

    screen_surface = SDL_GetWindowSurface(window);

    if (!screen_surface) {
        fprintf(stderr, "Error setting video mode: %s\n", SDL_GetError());
        return 1;
    } else {
        puts("SDL_GetWindowSurface ok");
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
 * Release graphics resources created by gfx_initialize.
 *
 * Frees the global screen surface and destroys the SDL window. Does not call SDL_Quit or free other SDL resources; callers should call SDL_Quit separately if needed.
 */
void gfx_finalize(void)
{
    SDL_FreeSurface(screen_surface);
    SDL_DestroyWindow(window);
}

/**
 * Set the global screen surface used for on-screen drawing.
 *
 * Assigns the provided SDL_Surface pointer to the internal screen surface
 * (used by functions like gfx_flip, gfx_putpixel_screen, and the *_screen
 * drawing wrappers). Passing NULL clears the current screen surface.
 *
 * @param screen SDL_Surface pointer to use as the active screen (or NULL).
 */
void gfx_set_screen_surface(SDL_Surface *screen)
{
    screen_surface = screen;
}

/**
 * Return the active screen surface.
 *
 * Returns the pointer to the SDL_Surface used as the current window/screen surface.
 * May be NULL if the graphics subsystem has not been initialized or the surface was not set.
 * @return SDL_Surface* Pointer to the screen surface, or NULL if unavailable.
 */
SDL_Surface* gfx_get_screen_surface(void)
{
    return screen_surface;
}

/**
 * Create a new 32-bit SDL surface with standard RGB masks.
 *
 * The surface uses a 32-bit pixel format with masks: R=0x00ff0000, G=0x0000ff00, B=0x000000ff,
 * and no alpha mask. The surface is created as an SDL system memory surface.
 *
 * @param width Width of the surface in pixels.
 * @param height Height of the surface in pixels.
 * @return Pointer to the newly created SDL_Surface on success, or NULL on failure.
 *         The caller is responsible for freeing the surface with SDL_FreeSurface().
 */
SDL_Surface* gfx_create_surface(int width, int height)
{
    return SDL_CreateRGBSurface(SDL_SWSURFACE, width, height, 32, 0x00ff0000, 0x0000ff00, 0x000000ff, 0x00000000);
}

/**
 * Blit an entire surface onto the global screen surface.
 *
 * Copies the full contents of `surface` to the application's main screen surface.
 * Note: the `x` and `y` parameters are currently ignored and the source is blitted
 * to the top-left of the destination.
 *
 * @param surface Source SDL_Surface whose whole area will be copied.
 * @param x Horizontal destination coordinate (currently unused; blit uses origin).
 * @param y Vertical destination coordinate (currently unused; blit uses origin).
 */
void gfx_bitblt(SDL_Surface *surface, const int x, const int y)
{
    SDL_BlitSurface(surface, NULL, screen_surface, NULL);
}

/**
 * Update the application window with the current screen surface.
 *
 * Copies the contents of the configured screen surface to the visible window
 * by calling SDL_UpdateWindowSurface. No return value; requires the window
 * to be initialized. */
void gfx_flip()
{
    SDL_UpdateWindowSurface(window);
}

/**
 * Fill the active screen surface with a uniform color.
 *
 * The color value is a pixel value encoded for the screen surface's pixel format
 * (Uint32). This fills the global screen surface (screen_surface) entirely.
 *
 * @param color Pixel color to fill the screen with ( Uint32, in the surface's pixel format ).
 */
void gfx_clear_screen(Uint32 color)
{
    SDL_FillRect(screen_surface, NULL, color);
}

/**
 * Retrieve the color of the pixel at (x, y) from an SDL surface.
 *
 * Reads the surface pixel at coordinates (x, y) and writes its red, green,
 * and blue components (0–255) into the locations pointed to by `r`, `g`,
 * and `b`, respectively. Only 24-bit and 32-bit surfaces are supported;
 * other pixel formats are ignored. If (x, y) is outside the surface bounds,
 * the function does nothing.
 *
 * @param surface SDL_Surface to read from.
 * @param x Horizontal pixel coordinate.
 * @param y Vertical pixel coordinate.
 * @param r Output pointer for the red component (must be non-NULL).
 * @param g Output pointer for the green component (must be non-NULL).
 * @param b Output pointer for the blue component (must be non-NULL).
 */
void gfx_getpixel(SDL_Surface *surface, int x, int y, unsigned char *r, unsigned char *g, unsigned char *b) {
    if (x>=0 && x< surface->w && y>=0 && y < surface->h) {
        if (surface->format->BitsPerPixel == 24) {
            Uint8 *pixel = (Uint8 *)surface->pixels;
            pixel += x*3;
            pixel += y*surface->pitch;
            *b = *pixel++;
            *g = *pixel++;
            *r = *pixel;
        }
        if (surface->format->BitsPerPixel == 32) {
            Uint8 *pixel = (Uint8 *)surface->pixels;
            pixel += x*4;
            pixel += y*surface->pitch;
            *b = *pixel++;
            *g = *pixel++;
            *r = *pixel;
        }
    }
}

/**
 * Set a pixel on an SDL surface to an RGB color.
 *
 * Writes the specified color to pixel (x,y) on the given SDL_Surface if the
 * coordinates are inside the surface bounds. Supports 24-bit and 32-bit
 * surfaces; colors are written in B,G,R byte order to match the surface
 * memory layout. If (x,y) is out of bounds or the surface has an unsupported
 * BitsPerPixel, the function does nothing.
 *
 * @param surface Destination SDL_Surface.
 * @param x X coordinate (0 = left).
 * @param y Y coordinate (0 = top).
 * @param r Red component (0–255).
 * @param g Green component (0–255).
 * @param b Blue component (0–255).
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
 * Plot a pixel on the active screen surface at the given coordinates using an RGB color.
 *
 * Coordinates outside the screen surface bounds are ignored.
 *
 * @param x X coordinate (pixels).
 * @param y Y coordinate (pixels).
 * @param r Red component (0-255).
 * @param g Green component (0-255).
 * @param b Blue component (0-255).
 */
void gfx_putpixel_screen(int x, int y, unsigned char r, unsigned char g, unsigned char b)
{
    gfx_putpixel(screen_surface, x, y, r, g, b);
}

/**
 * Draw a horizontal line on a surface.
 *
 * Draws a horizontal line from the lesser of `x1`/`x2` to the greater at row `y`,
 * setting each pixel to the specified RGB color.
 *
 * @param surface Destination SDL surface to draw into.
 * @param x1 One endpoint x-coordinate.
 * @param x2 Other endpoint x-coordinate.
 * @param y Y-coordinate (row) where the line will be drawn.
 * @param r Red component (0-255).
 * @param g Green component (0-255).
 * @param b Blue component (0-255).
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
 * Draws a vertical line of color (r,g,b) at column `x` between `y1` and `y2` (inclusive).
 * The endpoints may be provided in either order; the function draws from min(y1,y2) to max(y1,y2).
 * Pixels outside the surface bounds are ignored.
 *
 * @param surface Destination SDL surface.
 * @param x X coordinate (column) where the line is drawn.
 * @param y1 One end Y coordinate.
 * @param y2 Other end Y coordinate.
 * @param r Red component (0–255).
 * @param g Green component (0–255).
 * @param b Blue component (0–255).
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
 * Draw a straight (Bresenham) line between two points on a surface.
 *
 * Renders a 1-pixel-wide line from (x1, y1) to (x2, y2) onto the given SDL surface
 * using an integer Bresenham-like algorithm. The color is specified by RGB
 * components (0–255). Endpoints are included. Pixel bounds are not checked
 * here — out-of-bounds writes are handled (silently ignored) by gfx_putpixel.
 *
 * @param surface Destination SDL_Surface to draw onto.
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
 * Draw an anti-aliased line onto an SDL surface.
 *
 * Renders a smoothly blended line between (x1,y1) and (x2,y2) on the provided
 * SDL_Surface using a simple fractional-intensity algorithm (fixed-point
 * stepping with 256 subunits). Vertical and horizontal lines are delegated to
 * the non-anti-aliased hline/vline primitives for efficiency. The routine
 * writes two pixels per step with complementary intensities to approximate
 * anti-aliasing.
 *
 * Parameters:
 * - surface: destination SDL_Surface (must be valid and writable).
 * - x1,y1: coordinates of the start point.
 * - x2,y2: coordinates of the end point.
 * - r,g,b: 8-bit color components for the line.
 *
 * Notes:
 * - The function assumes a valid surface and does not perform NULL checks.
 * - Color components are applied directly; the implementation mixes two
 *   intensities of the provided color to produce the anti-aliased effect.
 * - If the line is perfectly horizontal or vertical, this function forwards
 *   to the corresponding non-anti-aliased primitive.
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
 * Draw a horizontal line on the current screen surface.
 *
 * Draws a horizontal line at row `y` from min(x1, x2) to max(x1, x2) (inclusive)
 * on the global screen surface using the specified RGB color components.
 *
 * @param x1 One endpoint X coordinate.
 * @param x2 Other endpoint X coordinate.
 * @param y  Y coordinate of the line.
 * @param r  Red component (0-255).
 * @param g  Green component (0-255).
 * @param b  Blue component (0-255).
 */
void gfx_hline_screen(int x1, int x2, int y, unsigned char r, unsigned char g, unsigned char b)
{
    gfx_hline(screen_surface, x1, x2, y, r, g, b);
}

/**
 * Draw a vertical line on the active screen surface.
 *
 * Draws a vertical line at column `x` from row `y1` to `y2` on the global screen surface using the
 * RGB color (`r`,`g`,`b`). Coordinates outside the surface bounds are ignored.
 *
 * @param x Column (horizontal) coordinate of the line.
 * @param y1 Starting row (vertical) coordinate.
 * @param y2 Ending row (vertical) coordinate.
 * @param r Red color component (0–255).
 * @param g Green color component (0–255).
 * @param b Blue color component (0–255).
 */
void gfx_vline_screen(int x, int y1, int y2, unsigned char r, unsigned char g, unsigned char b)
{
    gfx_vline(screen_surface, x, y1, y2, r, g, b);
}

/**
 * Draw a straight line on the current screen surface.
 *
 * Draws a line from (x1, y1) to (x2, y2) on the global screen surface using the
 * specified RGB color components (0–255). Coordinates are in pixels; pixels
 * outside the surface bounds are ignored.
 *
 * @param x1 X coordinate of the start point.
 * @param y1 Y coordinate of the start point.
 * @param x2 X coordinate of the end point.
 * @param y2 Y coordinate of the end point.
 * @param r Red component (0–255).
 * @param g Green component (0–255).
 * @param b Blue component (0–255).
 */
void gfx_line_screen(int x1, int y1, int x2, int y2, unsigned char r, unsigned char g, unsigned char b)
{
    gfx_line(screen_surface, x1, y1, x2, y2, r, g, b);
}

/**
 * Draw an anti-aliased line on the current screen surface.
 *
 * Renders an anti-aliased line between (x1, y1) and (x2, y2) using the
 * provided RGB color on the global screen surface managed by the gfx module.
 *
 * @param x1 X coordinate of the start point.
 * @param y1 Y coordinate of the start point.
 * @param x2 X coordinate of the end point.
 * @param y2 Y coordinate of the end point.
 * @param r  Red component of the line color (0-255).
 * @param g  Green component of the line color (0-255).
 * @param b  Blue component of the line color (0-255).
 */
void gfx_aa_line_screen(int x1, int y1, int x2, int y2, unsigned char r, unsigned char g, unsigned char b)
{
    gfx_aa_line(screen_surface, x1, y1, x2, y2, r, g, b);
}

/**
 * Set the bitmap font sprite-sheet used by bitmap font rendering functions.
 *
 * The provided SDL_Surface becomes the source image for gfx_print_char_bitmap_font
 * and gfx_print_string_bitmap_font. Passing NULL disables bitmap-font drawing.
 *
 * @param surface SDL_Surface pointer to the bitmap font sprite sheet (or NULL).
 */
void gfx_set_bitmap_font_surface(SDL_Surface *surface)
{
    bitmap_font_surface = surface;
}

/**
 * Render a single ASCII character from the bitmap font sprite sheet onto a surface.
 *
 * Renders character `ch` at destination coordinates (x, y) on `surface` using the
 * globally configured bitmap font sprite sheet (set via gfx_set_bitmap_font_surface()).
 * ASCII codes below 32 (control characters) are ignored; the function returns without
 * drawing in that case. Glyphs are assumed arranged vertically in the sprite sheet
 * with fixed width BITMAP_FONT_CHARACTER_WIDTH and height BITMAP_FONT_CHARACTER_HEIGHT.
 *
 * @param surface Destination SDL_Surface to blit the glyph onto.
 * @param x X coordinate on `surface` where the glyph's top-left will be placed.
 * @param y Y coordinate on `surface` where the glyph's top-left will be placed.
 * @param ch ASCII character to render.
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
 * Render a null-terminated string using the bitmap font sprite sheet.
 *
 * Renders each character in `str` onto `surface` starting at position (x, y).
 * Characters are drawn by calling gfx_print_char_bitmap_font and the x position
 * is advanced by BITMAP_FONT_CHARACTER_WIDTH after each character.
 *
 * The bitmap font surface must have been set with gfx_set_bitmap_font_surface()
 * prior to calling this function. `str` must be a non-NULL, null-terminated C
 * string.
 *
 * @param surface Destination SDL surface to draw the string onto (e.g., screen).
 * @param x X coordinate of the top-left position for the first character.
 * @param y Y coordinate of the top-left position for the first character.
 * @param str Null-terminated string to render (must not be NULL).
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
 * Render a NUL-terminated string onto the global screen surface using the bitmap font.
 *
 * Draws characters from the previously set bitmap font sprite sheet at pixel coordinates (x,y).
 * Each character is rendered with the module's fixed bitmap font cell size; the x position is
 * advanced by BITMAP_FONT_CHARACTER_WIDTH for each character. If no bitmap font surface has
 * been set, the call has no effect.
 *
 * @param x X coordinate (pixels) of the top-left corner where the first character is drawn.
 * @param y Y coordinate (pixels) of the top-left corner where the first character is drawn.
 * @param str NUL-terminated C string to render.
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

