#include <stdio.h>
#include <stdlib.h>
#include <math.h>
#include <SDL2/SDL.h>

#include "gfx.h"

#define WIDTH 320
#define HEIGHT 240

#define nil NULL

SDL_Surface *pixmap;

double xpos = -0.75;
double ypos = 0.0;
double scale = 240.0;
double  uhel = 45.0;

/**
 * Initialize the graphics subsystem (gfx/SDL) for the program.
 *
 * Attempts to initialize the gfx layer with a 640x480, 32-bit display.
 * On failure the function terminates the process with exit status 1.
 */
static void init_sdl(void)
{
    int result = gfx_initialize(0, 640, 480, 32);
    if (result)
    {
        exit(1);
    }
}

/**
 * Shut down the graphics subsystem and clean up SDL.
 *
 * Performs final cleanup by calling gfx_finalize() to terminate the gfx layer
 * and SDL_Quit() to shut down the SDL library. */
void finalize(void)
{
    gfx_finalize();
    SDL_Quit();
}

/**
 * Present the given pixmap surface on the screen.
 *
 * Blits the provided surface to the display origin and flips the screen buffer
 * to make the drawn content visible.
 *
 * @param surface SDL surface containing the image to present (origin at 0,0).
 */
static void show_fractal(SDL_Surface *surface)
{
    gfx_bitblt(surface, 0, 0);
    gfx_flip();
}

/**
 * Compute the visible fractal-plane bounds for the current view.
 *
 * Given the view center (xpos, ypos) and a zoom factor (scale), sets the
 * output parameters to the minimum and maximum X/Y coordinates visible
 * on the drawing surface.
 *
 * @param xpos Center X coordinate in fractal coordinate space.
 * @param ypos Center Y coordinate in fractal coordinate space.
 * @param scale Zoom factor (larger = zoomed in); used to convert screen
 *              pixels to fractal-plane units.
 * @param xmin Output pointer receiving the left (minimum X) bound.
 * @param ymin Output pointer receiving the top  (minimum Y) bound.
 * @param xmax Output pointer receiving the right (maximum X) bound.
 * @param ymax Output pointer receiving the bottom (maximum Y) bound.
 */
void calcCorner(double xpos, double ypos, double scale,
                double *xmin,  double *ymin,  double *xmax, double *ymax)
{
    *xmin=xpos-WIDTH/scale;
    *ymin=ypos-HEIGHT/scale;
    *xmax=xpos+WIDTH/scale;
    *ymax=ypos+HEIGHT/scale;
}

/**
 * Fill a surface with white and draw a light grid overlay.
 *
 * The surface is cleared to white (0xffffffff) and vertical and horizontal
 * grid lines are drawn every 20 pixels using the color RGB(191,191,255).
 *
 * @param surface SDL surface to render the background and grid onto.
 */
void draw_grid(SDL_Surface *surface)
{
    int width = surface->w;
    int height = surface->h;
    int x, y;
    SDL_FillRect(surface, NULL, 0xffffffff);

    for (x=0; x<width; x+=20)
    {
        gfx_vline(surface, x, 0, height-1, 191, 191, 255);
    }
    for (y=0; y<height; y+=20)
    {
        gfx_hline(surface, 0, width-1, y, 191, 191, 255);
    }
}

/**
 * Render a Mandelbrot-like fractal variant into the provided surface.
 *
 * Renders a 320x240 fractal using the global view parameters (xpos, ypos, scale)
 * to map pixel coordinates into the complex plane. For each pixel this function
 * iterates the escape function (with a negated real-part step) up to 150
 * iterations, stops when the squared magnitude exceeds 4.0, and writes a 32-bit
 * RGB color derived from the iteration count.
 *
 * The output is written directly into the surface's pixel buffer and thus the
 * surface must be writable and laid out as 32-bit pixels (4 bytes per pixel).
 *
 * @param surface SDL_Surface to draw into (must be a writable 32-bit surface).
 */
void draw_fractal_(SDL_Surface *surface)
{
    int x, y;
    Uint8 *pixel = nil;
    double cx, cy;
    double xmin, ymin, xmax, ymax;

    calcCorner(xpos, ypos, scale, &xmin, &ymin, &xmax, &ymax);
    cy = ymin;
    for (y=0; y<240; y++)
    {
        cx = xmin;
        pixel = (Uint8 *)surface->pixels + (y + 128) * surface->pitch + 160*4;
        for (x=0; x<320; x++)
        {
            double zx = 0.0;
            double zy = 0.0;
            unsigned int i = 0;
            while (i < 150) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                zx = -fabs(zx);
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                zy = 2.0 * zx * zy + cy;
                zx = zx2 - zy2 + cx;
                i++;
            }
            {
                int r = i*2;
                int g = i*3;
                int b = i*5;

                *pixel++ = r;
                *pixel++ = g;
                *pixel++ = b;
                pixel++;
            }
            cx += (xmax-xmin)/WIDTH;
        }
        cy += (ymax-ymin)/HEIGHT;
    }
}

/**
 * Render a Julia-set fractal into the given surface.
 *
 * Draws a 320×240 Julia fractal using the global view state (xpos, ypos, scale)
 * and the hard-coded complex parameter (ccx = 0.285, ccy = 0.01). For each pixel
 * the function iterates z <- z^2 + c up to 255 iterations, with a bailout when
 * |z| > 2. The pixel color is derived from the escape iteration count.
 *
 * The surface is written in-place; it is expected to be at least 320×240 and
 * use a 4-byte-per-pixel layout compatible with the renderer's pixel writes.
 */
void draw_fractal_julia(SDL_Surface *surface)
{
    int x, y;
    Uint8 *pixel = nil;
    double cx, cy;
    double xmin, ymin, xmax, ymax;
    double ccx, ccy;

    calcCorner(xpos, ypos, scale, &xmin, &ymin, &xmax, &ymax);
    /*
    ccx = 0.0;
    ccy = 1.0;
    */
    ccx = 0.285;
    ccy = 0.01;
    cy = ymin;
    for (y=0; y<240; y++)
    {
        cx = xmin;
        pixel = (Uint8 *)surface->pixels + (y + 128) * surface->pitch + 160*4;
        for (x=0; x<320; x++)
        {
            double zx = cx;
            double zy = cy;
            unsigned int i = 0;
            while (i < 255) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                /*zx = -fabs(zx);*/
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                zy = 2.0 * zx * zy + ccy;
                zx = zx2 - zy2 + ccx;
                i++;
            }
            {
                int r = i*2;
                int g = i*3;
                int b = i*5;

                *pixel++ = r;
                *pixel++ = g;
                *pixel++ = b;
                pixel++;
            }
            cx += (xmax-xmin)/WIDTH;
        }
        cy += (ymax-ymin)/HEIGHT;
    }
}

/**
 * Render a hybrid Julia/Mandelbrot fractal into the provided SDL surface.
 *
 * Iterates over the fixed viewport (WIDTH x HEIGHT) and for each pixel performs
 * a two-step iteration: first a Julia-like update using a fixed complex
 * constant, then a Mandelbrot-like update using the pixel's mapped coordinate
 * as the constant. Iteration stops at a maximum of 255 iterations or when the
 * escape radius (|z| > 2) is exceeded. The iteration count is mapped to an RGB
 * color and written directly into the surface pixels.
 *
 * The visible complex-plane bounds are computed from the global center (xpos,
 * ypos) and zoom (scale) via calcCorner. The function writes pixels in-place
 * into the provided SDL_Surface.
 *
 * @param surface SDL surface to draw the fractal onto (must be a writable
 *                surface with expected pitch and pixel format).
 */
void draw_fractal_julia_mandelbrot(SDL_Surface *surface)
{
    int x, y;
    Uint8 *pixel = nil;
    double cx, cy;
    double xmin, ymin, xmax, ymax;
    double ccx, ccy;

    calcCorner(xpos, ypos, scale, &xmin, &ymin, &xmax, &ymax);
    ccx = 0.285;
    ccy = 0.01;
    ccx = -0.7269;
    ccy = 0.1889;
    ccx = -1.0;
    ccy = 0.0;
    cy = ymin;
    for (y=0; y<240; y++)
    {
        cx = xmin;
        pixel = (Uint8 *)surface->pixels + (y + 128) * surface->pitch + 160*4;
        for (x=0; x<320; x++)
        {
            double zx = cx;
            double zy = cy;
            unsigned int i = 0;
            while (i < 255) {
                double zx2, zy2;

                zx2 = zx * zx;
                zy2 = zy * zy;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                zy = 2.0 * zx * zy + ccy;
                zx = zx2 - zy2 + ccx;
                i++;

                zx2 = zx * zx;
                zy2 = zy * zy;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                zy = 2.0 * zx * zy + cy;
                zx = zx2 - zy2 + cx;
                i++;
            }
            {
                int r = i*2;
                int g = i*3;
                int b = i*5;

                *pixel++ = r;
                *pixel++ = g;
                *pixel++ = b;
                pixel++;
            }
            cx += (xmax-xmin)/WIDTH;
        }
        cy += (ymax-ymin)/HEIGHT;
    }
}

/**
 * Render a multifractal image combining Mandelbrot- and Julia-like updates into the given surface.
 *
 * This draws a 320x240 fractal into `surface` by mapping the current global view (xpos, ypos, scale)
 * to complex-plane coordinates and iterating a hybrid update: for the first ~50 iterations each
 * pixel uses the pixel-mapped constant (Mandelbrot-like), then switches to a fixed constant
 * pair (Julia-like). Iteration stops when |z|^2 > 4.0 or after 255 iterations. The resulting
 * iteration count is written directly into the surface pixels as RGB using simple linear
 * multipliers (R=i*2, G=i*3, B=i*5).
 *
 * The function modifies the pixel buffer of `surface` in-place and assumes a 32-bit surface with
 * at least WIDTH x HEIGHT pixels. It reads global state (xpos, ypos, scale) to compute the view
 * bounds.
 *
 * @param surface Target SDL_Surface whose pixels will be updated with the fractal image.
 */
void draw_multifractal_mandel_julia(SDL_Surface *surface)
{
    int x, y;
    Uint8 *pixel = nil;
    double cx, cy;
    double xmin, ymin, xmax, ymax;
    double ccx, ccy;

    calcCorner(xpos, ypos, scale, &xmin, &ymin, &xmax, &ymax);
    /*
    ccx = 0.0;
    ccy = 1.0;
    */
    ccx = 0.285;
    ccy = 0.01;
    ccx=-1.5;
    ccy=0.0;
    cy = ymin;
    for (y=0; y<240; y++)
    {
        cx = xmin;
        pixel = (Uint8 *)surface->pixels + (y + 128) * surface->pitch + 160*4;
        for (x=0; x<320; x++)
        {
            double zx = 0.0;
            double zy = 0.0;
            unsigned int i = 0;
            while (i < 255) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                /*if (i%2==0) {*/
                if (i>50 /*&& i<30*/) {
                    zy = 2.0 * zx * zy + ccy;
                    zx = zx2 - zy2 + ccx;
                } else {
                    zy = 2.0 * zx * zy + cy;
                    zx = zx2 - zy2 + cx;
                }
                i++;
            }
            {
                int r = i*2;
                int g = i*3;
                int b = i*5;

                *pixel++ = r;
                *pixel++ = g;
                *pixel++ = b;
                pixel++;
            }
            cx += (xmax-xmin)/WIDTH;
        }
        cy += (ymax-ymin)/HEIGHT;
    }
}

/**
 * Render a rotated Mandelbrot/Julia hybrid fractal into the provided surface.
 *
 * Draws a 320x240 fractal view into `surface` by mapping pixels to the complex
 * plane using the current global center (xpos,ypos), zoom (scale), and
 * rotation angle (uhel). For each pixel the routine applies up to 64
 * iterations of a complex quadratic map with a rotation-dependent offset and
 * writes an RGB color computed from the iteration count into the surface's
 * pixel buffer (assumes 32-bit surface layout where three bytes are R,G,B
 * followed by an alpha/unused byte).
 *
 * @param surface SDL surface whose pixels will be modified with the rendered fractal.
 */
void draw_mandeljulia(SDL_Surface *surface)
{
    double  zx,zy,zx2,zy2,cx,cy;
    double  cosu,sinu,ccxc,ccyc;
    int     x,y,i;
    Uint8 *pixel = nil;

    double ccx = 0.0;
    double ccy = 0.0;

    double xmin, ymin, xmax, ymax;
    double u;

    calcCorner(xpos, ypos, scale, &xmin, &ymin, &xmax, &ymax);

    u=uhel*3.1415/180.0;
    cosu=cos(u);
    sinu=sin(u);
    ccxc=ccx*cosu;
    ccyc=ccy*cosu;

    cy = ymin;

    for (y=0;y<240;y++) {
        cx=xmin;
        pixel = (Uint8 *)surface->pixels + (y + 128) * surface->pitch + 160*4;
        for (x=0;x<320;x++) {
            i=0;
            zx=cx*cosu;
            zy=cy*cosu;
            do {
                zx2=zx*zx;
                zy2=zy*zy;
                zy=2.0*zx*zy+ccyc+cy*sinu;
                zx=zx2-zy2+ccxc+cx*sinu;
                i++;
            } while (i<64 && zx2+zy2<4.0);
            {
                int r = i*2;
                int g = i*3;
                int b = i*5;

                *pixel++ = r;
                *pixel++ = g;
                *pixel++ = b;
                pixel++;
            }
            cx += (xmax-xmin)/WIDTH;
        }
        cy += (ymax-ymin)/HEIGHT;
    }
}

/**
 * Render a multifractal into the provided SDL surface.
 *
 * Renders a 320×240 multifractal by iterating a complex quadratic map that alternates
 * between two parameter pairs depending on the iteration count. The visible region is
 * computed from the global xpos, ypos and scale variables. Iteration stops when the
 * squared magnitude exceeds 4.0 or when 255 iterations are reached. The resulting
 * iteration count is mapped to RGB using simple linear multipliers and written
 * directly into the surface pixel buffer (32-bit RGBA layout).
 *
 * @param surface Destination SDL_Surface whose pixel buffer will be written. The
 *                function expects a surface matching the program's drawing area
 *                (320×240, 32 bpp) so that pitch and pixel layout match the writer.
 */
void draw_multifractal(SDL_Surface *surface)
{
    int x, y;
    Uint8 *pixel = nil;
    double cx, cy;
    double xmin, ymin, xmax, ymax;
    double ccx1, ccy1;
    double ccx2, ccy2;

    calcCorner(xpos, ypos, scale, &xmin, &ymin, &xmax, &ymax);
    ccx1 = 0.0;
    ccy1 = 1.0;
    ccx2 = -1.5;
    ccy2 = 0.0;
    cy = ymin;
    for (y=0; y<240; y++)
    {
        cx = xmin;
        pixel = (Uint8 *)surface->pixels + (y + 128) * surface->pitch + 160*4;
        for (x=0; x<320; x++)
        {
            double zx = cx;
            double zy = cy;
            unsigned int i = 0;
            while (i < 255) {
                double zx2 = zx * zx;
                double zy2 = zy * zy;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                if (i>20) {
                    zy = 2.0 * zx * zy + ccy1;
                    zx = zx2 - zy2 + ccx1;
                } else {
                    zy = 2.0 * zx * zy + ccy2;
                    zx = zx2 - zy2 + ccx2;
                }
                i++;
            }
            {
                int r = i*2;
                int g = i*3;
                int b = i*5;

                *pixel++ = r;
                *pixel++ = g;
                *pixel++ = b;
                pixel++;
            }
            cx += (xmax-xmin)/WIDTH;
        }
        cy += (ymax-ymin)/HEIGHT;
    }
}

/**
 * Render a Mandelbrot-like fractal into the provided SDL surface.
 *
 * This routine maps the current view (centered at globals `xpos`,`ypos` with zoom `scale`)
 * onto the surface and computes a per-pixel escape-time fractal. Each pixel is iterated
 * with two alternating complex parameter pairs (ccx1/ccy1 and ccx2/ccy2) until the
 * magnitude exceeds 2.0 (squared magnitude > 4.0) or an iteration limit (255) is reached.
 * The final iteration count is converted to an RGB color and written directly into the
 * surface pixel buffer.
 *
 * @param surface SDL surface to draw into; pixels are written in-place.
 */
void draw_fractal(SDL_Surface *surface)
{
    int x, y;
    Uint8 *pixel = nil;
    double cx, cy;
    double xmin, ymin, xmax, ymax;
    /* not used for Mandelbrot fractal types
    double ccx, ccy;
    */
    double ccx1, ccx2, ccy1, ccy2;

    calcCorner(xpos, ypos, scale, &xmin, &ymin, &xmax, &ymax);
    /* not used for Mandelbrot fractal types
    ccx = -0.7269;
    ccy = 0.1889;
    */
    ccx1 = 0.0;
    ccy1 = 1.0;
    ccx2 = 0.285;
    ccy2 = 0.01;
    cy = ymin;
    for (y=0; y<240; y++)
    {
        cx = xmin;
        pixel = (Uint8 *)surface->pixels + (y + 128) * surface->pitch + 160*4;
        for (x=0; x<320; x++)
        {
            double zx = cx;
            double zy = cy;
            unsigned int i = 0;
            while (i < 255) {
                double zx2, zy2;

                zx2 = zx * zx;
                zy2 = zy * zy;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                zy = 2.0 * zx * zy + ccy1;
                zx = zx2 - zy2 + ccx1;
                i++;

                zx2 = zx * zx;
                zy2 = zy * zy;
                if (zx2 + zy2 > 4.0) {
                    break;
                }
                zy = 2.0 * zx * zy + ccy2;
                zx = zx2 - zy2 + ccx2;
                i++;
            }
            {
                int r = i*2;
                int g = i*3;
                int b = i*5;

                *pixel++ = r;
                *pixel++ = g;
                *pixel++ = b;
                pixel++;
            }
            cx += (xmax-xmin)/WIDTH;
        }
        cy += (ymax-ymin)/HEIGHT;
    }
}

/**
 * Redraw the given pixmap: draw the background grid, render the Mandel–Julia fractal, and present it.
 *
 * @param pixmap SDL surface used as the drawing buffer; contents will be updated and then copied to the screen.
 */
void redraw(SDL_Surface *pixmap)
{
    draw_grid(pixmap);
    draw_mandeljulia(pixmap);
    show_fractal(pixmap);
}

/**
 * Run the main SDL event loop to handle user input and update fractal view.
 *
 * Processes SDL events until quit (ESC, 'q', or window close). Keyboard controls:
 * - Arrow keys: pan the view (updates global xpos/ypos)
 * - PageUp / PageDown: zoom out / in (updates global scale)
 * - 'z' / 'x': rotate view (updates global uhel)
 *
 * When input changes the view, requests a redraw via redraw(pixmap); otherwise
 * it idles briefly (SDL_Delay). This function has no return value and operates
 * by mutating program-wide state used by the rendering code.
 */
static void main_event_loop(void)
{
    SDL_Event event;
    int done = 0;
    int left = 0, right = 0, up = 0, down = 0;
    int zoomin = 0, zoomout = 0;
    int perform_redraw;
    int angle_1 = 0, angle_2 = 0;

    do
    {
        /*SDL_WaitEvent(&event);*/
        while (SDL_PollEvent(&event))
        {
            switch (event.type)
            {
                case SDL_QUIT:
                    done = 1;
                    break;
                case SDL_KEYDOWN:
                    /* TODO - control: more speed */
                    /* TODO - shift: lower speed */
                    switch (event.key.keysym.sym)
                    {
                        case SDLK_ESCAPE:
                        case SDLK_q:
                            done = 1;
                            break;
                        case SDLK_LEFT:
                            left = 1;
                            break;
                        case SDLK_RIGHT:
                            right = 1;
                            break;
                        case SDLK_UP:
                            up = 1;
                            break;
                        case SDLK_DOWN:
                            down = 1;
                            break;
                        case SDLK_PAGEDOWN:
                            zoomin = 1;
                            break;
                        case SDLK_PAGEUP:
                            zoomout = 1;
                            break;
                        case SDLK_z:
                            angle_1 = 1;
                            break;
                        case SDLK_x:
                            angle_2 = 1;
                            break;
                        default:
                            break;
                    }
                    break;
                case SDL_KEYUP:
                    switch (event.key.keysym.sym)
                    {
                        case SDLK_LEFT:
                            left = 0;
                            break;
                        case SDLK_RIGHT:
                            right = 0;
                            break;
                        case SDLK_UP:
                            up = 0;
                            break;
                        case SDLK_DOWN:
                            down = 0;
                            break;
                        case SDLK_PAGEDOWN:
                            zoomin = 0;
                            break;
                        case SDLK_PAGEUP:
                            zoomout = 0;
                            break;
                        case SDLK_z:
                            angle_1 = 0;
                            break;
                        case SDLK_x:
                            angle_2 = 0;
                            break;
                        default:
                            break;
                    }
                default:
                    break;
            }
        }
        perform_redraw = 0;
        if (left) {
            xpos -= 10.0/scale;
            perform_redraw=1;
        }
        if (right) {
            xpos += 10.0/scale;
            perform_redraw=1;
        }
        if (up) {
            ypos -= 10.0/scale;
            perform_redraw=1;
        }
        if (down) {
            ypos += 10.0/scale;
            perform_redraw=1;
        }
        if (zoomin) {
            scale *= 0.9;
            perform_redraw=1;
        }
        if (zoomout) {
            scale *= 1.1;
            perform_redraw=1;
        }
        if (angle_1) {
            uhel--;
            perform_redraw=1;
        }
        if (angle_2) {
            uhel++;
            perform_redraw=1;
        }
        if (perform_redraw) {
            redraw(pixmap);
        } else {
            SDL_Delay(10);
        }
    } while (!done);
}

/**
 * Program entry point: initialize graphics, render initial fractal, run event loop, and clean up.
 *
 * Initializes the gfx/SDL subsystems, creates an offscreen pixmap sized to the screen,
 * draws the initial grid and Mandelbrot–Julia fractal, presents the result, then enters
 * the interactive event loop until the user requests exit and performs final cleanup.
 *
 * @param argc Number of command-line arguments (unused).
 * @param argv Command-line arguments (unused).
 * @return Exit status (0 on normal termination).
 */
int main(int argc, char **argv)
{
    /*SDL_Surface *font;*/
    SDL_Surface *screen;

    init_sdl();

    screen = gfx_get_screen_surface();
    pixmap = gfx_create_surface(screen->w, screen->h);

    draw_grid(pixmap);
    draw_mandeljulia(pixmap);
    show_fractal(pixmap);
    main_event_loop();
    finalize();
    return 0;
}
