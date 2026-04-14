import board
import digitalio
import time
import busio
import neopixel
import adafruit_ssd1306
import struct
import rotaryio
import usb_hid
from adafruit_hid.keyboard import Keyboard
from adafruit_hid.keycode import Keycode
from adafruit_framebuf import BitmapFont
from adafruit_hid.mouse import Mouse
from adafruit_hid.consumer_control import ConsumerControl
from adafruit_hid.consumer_control_code import ConsumerControlCode
from math import sqrt, cos, sin, radians


class FastBitmapFont(BitmapFont):
    def draw_char(self, char, x, y, framebuffer, color, size=1):
        # Only use fast path when 8-pixel aligned and size == 1
        if size != 1 or (y % 8) != 0:
            return super().draw_char(char, x, y, framebuffer, color, size)
        for cx in range(self.font_width):
            self._font.seek(2 + (ord(char) * self.font_width) + cx)
            try:
                line = struct.unpack("B", self._font.read(1))[0]
            except RuntimeError:
                continue
            framebuffer.buf[framebuffer.width * (y >> 3) + x + cx] |= line
            
# NeoPixel setup
np = neopixel.NeoPixel(board.GP5, 4, auto_write=False)

#HID
kbd = Keyboard(usb_hid.devices)

# I2C & OLED setup
i2c = busio.I2C(scl=board.GP1, sda=board.GP0, frequency=1000000)
display = adafruit_ssd1306.SSD1306_I2C(128, 64, i2c)
display._font = FastBitmapFont()

#create local framebuffer
#buf = bytearray((128 * 64) // 8)
#fb = framebuf.FrameBuffer(buf, 128, 64, framebuf.MONO_VLSB)

# Row driver outputs
rw1 = digitalio.DigitalInOut(board.GP9); rw1.direction = digitalio.Direction.OUTPUT
rw2 = digitalio.DigitalInOut(board.GP8); rw2.direction = digitalio.Direction.OUTPUT
rw3 = digitalio.DigitalInOut(board.GP7); rw3.direction = digitalio.Direction.OUTPUT
rw4 = digitalio.DigitalInOut(board.GP6); rw4.direction = digitalio.Direction.OUTPUT

# Column input with pull-down
cl1 = digitalio.DigitalInOut(board.GP4); cl1.direction = digitalio.Direction.INPUT; cl1.pull = digitalio.Pull.DOWN
cl2 = digitalio.DigitalInOut(board.GP3); cl2.direction = digitalio.Direction.INPUT; cl2.pull = digitalio.Pull.DOWN
cl3 = digitalio.DigitalInOut(board.GP2); cl3.direction = digitalio.Direction.INPUT; cl3.pull = digitalio.Pull.DOWN

# Encoder inputs & button with pull-up
encButton = digitalio.DigitalInOut(board.GP16); encButton.direction = digitalio.Direction.INPUT; encButton.pull = digitalio.Pull.UP

# On-board LED (RP2040)
led = digitalio.DigitalInOut(board.LED)
led.direction = digitalio.Direction.OUTPUT

# Initialize NeoPixel
np[1] = (0, 0, 0)  # full-bright magenta
np.write()

cc = ConsumerControl(usb_hid.devices)
buttons = [0] * 12 #cerry mx buttons, left to right, top to bottom starting at index 0
mouse = Mouse(usb_hid.devices)
enc = rotaryio.IncrementalEncoder(board.GP14, board.GP15)
last_position = None

last_activity = time.monotonic()
display_on = True
def reset_activity():
    global last_activity, display_on
    last_activity = time.monotonic()
    if not display_on:
        display.poweron()
        display_on = True
        
def check_display_timeout():
    global display_on
    if display_on and (time.monotonic() - last_activity > 3.0):
        display.fill(0)
        display.show()
        display.poweroff()
        display_on = False

def button_read():
    rw4.value = 1
    buttons[0] = cl1.value
    buttons[4] = cl2.value
    buttons[8] = cl3.value
    rw4.value = 0
    rw3.value = 1
    buttons[1] = cl1.value
    buttons[5] = cl2.value
    buttons[9] = cl3.value
    rw3.value = 0
    rw2.value = 1
    buttons[2] = cl1.value
    buttons[6] = cl2.value
    buttons[10] = cl3.value
    rw2.value = 0
    rw1.value = 1
    buttons[3] = cl1.value
    buttons[7] = cl2.value
    buttons[11] = cl3.value
    rw1.value = 0
#     print(buttons)
    
def snake_draw():
    global x_val, y_val
    if buttons[1]:
        y_val -=1
    if buttons[9]:
        y_val +=1
    if buttons[6]:
        x_val +=1
    if buttons[4]:
        x_val -=1
    if buttons[3]:
        display.fill(0)
    x_val = x_val % 128
    y_val = y_val % 64
    display.pixel(x_val,y_val,1)    

def snake_play():
    global x_val, y_val
    if buttons[1]:
        y_val -=1
    if buttons[9]:
        y_val +=1
    if buttons[6]:
        x_val +=1
    if buttons[4]:
        x_val -=1
    if buttons[3]:
        display.fill(0)
    x_val = x_val % 128
    y_val = y_val % 64
    #get_pixel(framebuffer, 1, 2)
    display.pixel(x_val,y_val,1)
    
rgb_val = 0
def rgb_iterate():
    global rgb_val
    if rgb_val > 360:
        rgb_val = 0
    var = palette[rgb_val]
    np[0] = var
    np[1] = var
    np[2] = var
    np[3] = var
    rgb_val += 2
    np.show()
    

        

led_select = 0
def led_mode_select():
    global prev_button
    global led_select
    if prev_button == 0 and buttons[10] == 1:
        led_select +=1
    if led_select > 2:
        led_select = 0
    prev_button = buttons[10]
    print(led_select)
    
def remap(v):
    # scale from [-1..1] to [0..255]
    return int(((255 * v + 85) * 0.75) + 0.5)

def rotate(deg):
    a = radians(deg)
    c = cos(a)
    s = sin(a)
    # Based on color-rotation math from FeatherWing example
    r = c + (1.0 - c) / 3.0
    g = (1.0 - c) / 3.0 + sqrt(1.0 / 3.0) * s
    b = (1.0 - c) / 3.0 - sqrt(1.0 / 3.0) * s
    return (remap(r), remap(g), remap(b))

# Precompute a full 360° palette
palette = [rotate(d) for d in range(360)]

offset = 0

def lang_select():
    global lang
    if buttons[1] and buttons[0]:
        lang +=1
    if lang > 1:
        lang = 0
        
prev_button = buttons[11]
select = 0
def menu_select():
    global prev_button, select
    if prev_button == 0 and buttons[11] == 1:
        select +=1
    if select > 5:
        select = 0
    prev_button = buttons[11]
move_amount = 20
def keybrd1():
    global time_prev_keyboard
    if buttons[1]:
        kbd.press(Keycode.W)
    else:
         kbd.release(Keycode.W)

    if buttons[4]:
        kbd.press(Keycode.A)
    else:
        kbd.release(Keycode.A)

    if buttons[5]:
        kbd.press(Keycode.S)
    else:
        kbd.release(Keycode.S)
        
    if buttons[6]:
        kbd.press(Keycode.D)
    else:
        kbd.release(Keycode.D)
        
    if buttons[9]:
        kbd.press(Keycode.SPACE)
    else:
        kbd.release(Keycode.SPACE)
        
    if buttons[0]:  #up
        mouse.move(y=-move_amount)
        reset_activity()
        
    if buttons[8]:  #down
        mouse.move(y=move_amount)
        reset_activity()
        
    if buttons[10]:  #left
        mouse.move(x=-move_amount)
        reset_activity()
        
    if buttons[2]:  #right
        mouse.move(x=move_amount)
        reset_activity()
        
    if time.monotonic_ns() - time_prev_keyboard > 100000:
        if buttons[3]:
            mouse.click(Mouse.LEFT_BUTTON)
            reset_activity()
            
        if buttons[7]:
            mouse.click(Mouse.RIGHT_BUTTON)
            reset_activity()
        
        
        time_prev_keyboard = time.monotonic_ns()
    
    global last_position
    position = enc.position
    if last_position is None:
        last_position = position
        return

    delta = position - last_position
    if delta != 0:
        mouse.move(wheel=-delta)
        reset_activity()

    last_position = position

def keybrd2():
    if buttons[2]:
        kbd.press(Keycode.CONTROL)
    else:
        kbd.release(Keycode.CONTROL)
        
    if buttons[0]:
        kbd.press(Keycode.SHIFT)
    else:
        kbd.release(Keycode.SHIFT)
        
    if buttons[1]:
        kbd.press(Keycode.E)
    else:
        kbd.release(Keycode.E)
        
    if buttons[3]:
        kbd.press(Keycode.ESCAPE)
    else:
        kbd.release(Keycode.ESCAPE)
        
    if buttons[4]:
        kbd.press(Keycode.BACKSPACE)
    else:
        kbd.release(Keycode.BACKSPACE)
        
    if buttons[5]:
        kbd.press(Keycode.Q)
    else:
        kbd.release(Keycode.Q)
        
    if buttons[6]:
        kbd.press(Keycode.ENTER)
    else:
        kbd.release(Keycode.ENTER)
    
    if buttons[7]:
        kbd.press(Keycode.CAPS_LOCK)
    else:
        kbd.release(Keycode.CAPS_LOCK)
        
    if buttons[8]:
        kbd.press(Keycode.C)
    else:
        kbd.release(Keycode.C)
        
    if buttons[9]:
        kbd.press(Keycode.V)
    else:
        kbd.release(Keycode.V)
        
    

        
    global last_position
    position = enc.position
    if last_position is None:
        last_position = position
    delta = position - last_position
    if delta > 0:
        cc.send(ConsumerControlCode.VOLUME_INCREMENT)
        
    elif delta < 0:
        cc.send(ConsumerControlCode.VOLUME_DECREMENT)
    last_position = position
time_prev_keyboard = time.monotonic_ns()

lang = 0


def keybrd3():
    global lang
    if buttons[0]:
        kbd.press(Keycode.A)
    else:
        kbd.release(Keycode.A)
        
    if buttons[1]:
        kbd.press(Keycode.B)
    else:
        kbd.release(Keycode.B)
        
    if buttons[2]:
        kbd.press(Keycode.C)
    else:
        kbd.release(Keycode.C)
        
    if buttons[3]:
        kbd.press(Keycode.D)
    else:
        kbd.release(Keycode.D)
        
    if buttons[4]:
        kbd.press(Keycode.E)
    else:
        kbd.release(Keycode.E)
        
    if buttons[5]:
        kbd.press(Keycode.F)
    else:
        kbd.release(Keycode.F)
        
    if buttons[6]:
        kbd.press(Keycode.G)
    else:
        kbd.release(Keycode.G)
        
    if buttons[7]:
        kbd.press(Keycode.H)
    else:
        kbd.release(Keycode.H)
        
    if buttons[8]:
        kbd.press(Keycode.I)
    else:
        kbd.release(Keycode.I)
        
    if buttons[9]:
        kbd.press(Keycode.J)
    else:
        kbd.release(Keycode.J)
        
    if buttons[10]:
        kbd.press(Keycode.K)
    else:
        kbd.release(Keycode.K)
        
    global last_position
    position = enc.position
    if last_position is None:
        last_position = position
    delta = position - last_position
    if delta > 0:
        cc.send(ConsumerControlCode.BRIGHTNESS_INCREMENT)
        
    elif delta < 0:
        cc.send(ConsumerControlCode.BRIGHTNESS_DECREMENT)
    last_position = position
time_prev_keyboard = time.monotonic_ns()
    
        
def keybrd4():
    if buttons[0]:
        kbd.press(Keycode.L)
    else:
        kbd.release(Keycode.L)
        
    if buttons[1]:
        kbd.press(Keycode.M)
    else:
        kbd.release(Keycode.M)
        
    if buttons[2]:
        kbd.press(Keycode.N)
    else:
        kbd.release(Keycode.N)
        
    if buttons[3]:
        kbd.press(Keycode.O)
    else:
        kbd.release(Keycode.O)
        
    if buttons[4]:
        kbd.press(Keycode.P)
    else:
        kbd.release(Keycode.P)
        
    if buttons[5]:
        kbd.press(Keycode.Q)
    else:
        kbd.release(Keycode.Q)
        
    if buttons[6]:
        kbd.press(Keycode.R)
    else:
        kbd.release(Keycode.R)
        
    if buttons[7]:
        kbd.press(Keycode.S)
    else:
        kbd.release(Keycode.S)
        
    if buttons[8]:
        kbd.press(Keycode.T)
    else:
        kbd.release(Keycode.T)
        
    if buttons[9]:
        kbd.press(Keycode.U)
    else:
        kbd.release(Keycode.U)
        
    if buttons[10]:
        kbd.press(Keycode.V)
    else:
        kbd.release(Keycode.V)

def keybrd5():
    if buttons[0]:
        kbd.press(Keycode.W)
    else:
        kbd.release(Keycode.W)
        
    if buttons[1]:
        kbd.press(Keycode.X)
    else:
        kbd.release(Keycode.X)
    if lang == 1:#czech
        if buttons[2]:
            kbd.press(Keycode.Z)
        else:
            kbd.release(Keycode.Z)
        
        if buttons[3]:
            kbd.press(Keycode.Y)
        else:
           kbd.release(Keycode.Y)
    if lang == 0:#english
        if buttons[2]:
            kbd.press(Keycode.Y)
        else:
            kbd.release(Keycode.Y)
        
        if buttons[3]:
            kbd.press(Keycode.Z)
        else:
           kbd.release(Keycode.Z)
           
def keybrd6():
    global lang
    if lang == 0:
        if buttons[0]:
            kbd.press(Keycode.ZERO)
        else:
            kbd.release(Keycode.ZERO)
            
        if buttons[1]:
            kbd.press(Keycode.ONE)
        else:
            kbd.release(Keycode.ONE)
            
        if buttons[2]:
            kbd.press(Keycode.TWO)
        else:
            kbd.release(Keycode.TWO)
            
        if buttons[3]:
            kbd.press(Keycode.THREE)
        else:
            kbd.release(Keycode.THREE)
            
        if buttons[4]:
            kbd.press(Keycode.FOUR)
        else:
            kbd.release(Keycode.FOUR)
            
        if buttons[5]:
            kbd.press(Keycode.FIVE)
        else:
            kbd.release(Keycode.FIVE)
            
        if buttons[6]:
            kbd.press(Keycode.SIX)
        else:
            kbd.release(Keycode.SIX)
            
        if buttons[7]:
            kbd.press(Keycode.SEVEN)
        else:
            kbd.release(Keycode.SEVEN)
            
        if buttons[8]:
            kbd.press(Keycode.EIGHT)
        else:
            kbd.release(Keycode.EIGHT)
            
        if buttons[9]:
            kbd.press(Keycode.NINE)
        else:
            kbd.release(Keycode.NINE)
            
    if lang == 1:
        if buttons[0]:
            kbd.press(Keycode.SHIFT, Keycode.ZERO)
        else:
            kbd.release(Keycode.SHIFT, Keycode.ZERO)
            
        if buttons[1]:
            kbd.press(Keycode.SHIFT, Keycode.ONE)
        else:
            kbd.release(Keycode.SHIFT, Keycode.ONE)
            
        if buttons[2]:
            kbd.press(Keycode.SHIFT, Keycode.TWO)
        else:
            kbd.release(Keycode.SHIFT, Keycode.TWO)
            
        if buttons[3]:
            kbd.press(Keycode.SHIFT, Keycode.THREE)
        else:
            kbd.release(Keycode.SHIFT, Keycode.THREE)
            
        if buttons[4]:
            kbd.press(Keycode.SHIFT, Keycode.FOUR)
        else:
            kbd.release(Keycode.SHIFT, Keycode.FOUR)
            
        if buttons[5]:
            kbd.press(Keycode.SHIFT, Keycode.FIVE)
        else:
            kbd.release(Keycode.SHIFT, Keycode.FIVE)
            
        if buttons[6]:
            kbd.press(Keycode.SHIFT, Keycode.SIX)
        else:
            kbd.release(Keycode.SHIFT, Keycode.SIX)
            
        if buttons[7]:
            kbd.press(Keycode.SHIFT, Keycode.SEVEN)
        else:
            kbd.release(Keycode.SHIFT, Keycode.SEVEN)
            
        if buttons[8]:
            kbd.press(Keycode.SHIFT, Keycode.EIGHT)
        else:
            kbd.release(Keycode.SHIFT, Keycode.EIGHT)
            
        if buttons[9]:
            kbd.press(Keycode.SHIFT, Keycode.NINE)
        else:
            kbd.release(Keycode.SHIFT, Keycode.NINE)

    global last_position
    position = enc.position
    if last_position is None:
        last_position = position
        return

    delta = position - last_position
    if delta != 0:
        mouse.move(wheel=-delta)
        reset_activity()

    last_position = position
        
display.invert = False
display.contrast = 1

prev_encoder_a = 0
prev_encoder_b = 0
enc_rot = 0
x_val = 0
y_val = 0

cc.send(ConsumerControlCode.VOLUME_INCREMENT)

while True:
    #display.fill(0)
    #display.text("OLED 128x64", 80, 24, 1)
    #time.sleep(0.01)  
    button_read()
    #now = time.monotonic_ns()
    check_display_timeout()
    #print((time.monotonic_ns()-now)/1000000)
    if any(buttons):
        reset_activity()
    #display.scroll(1, 1)
    display.show()
    menu_select()
    caps_lock_on = kbd.led_on(Keyboard.LED_CAPS_LOCK)
    lang_select()
    if select == 0:
        keybrd1()
        display.fill(0)
        display.text("MOVEMENT + SCROLL",18,35,1)
        
    if select == 1:
        keybrd2()
        display.fill(0)
        display.text("FUNCTIONS + VOLUME",13,35,1)
        
    if select == 2:
        keybrd3()
        display.fill(0)
        display.text("A - K + BRIGHTNESS", 13,35,1)
        
    if select == 3:
        keybrd4()
        display.fill(0)
        display.text("L - V + LED", 35,35,1)
        
    if select == 4:
        keybrd5()
        display.fill(0)
        display.text("W - Z", 50,35,1)
        
    if select == 5:
        keybrd6()
        display.fill(0)
        display.text("NUMBERS 0 - 9", 25,35,1)
        
    if caps_lock_on:
        display.text("CAPS LOCK",0,10,1)
        
    if lang == 1:
        display.text("CZ", 0, 0, 1)
        
    if lang == 0:
        display.text("EN", 0, 0, 1)
   
    display.text("LED MODE:", 75, 0, 1)
        
    if led_select == 0:
        display.text("AUTO", 100, 10, 1)
        
    if led_select == 1:
        display.text("MANUAL", 95, 10, 1)
        
    if led_select == 2:
        display.text("OFF", 105, 10, 1)
        
    if select == 1:
        led_mode_select()
        
    if select == 3 and led_select == 1:
        np[0] = palette[enc.position*2%360]
        np[1] = palette[enc.position*2%360]
        np[2] = palette[enc.position*2%360]
        np[3] = palette[enc.position*2%360]
        np.write()
        
    if select == 1:
        led_mode_select()
        
    if led_select == 0:
        rgb_iterate()
        
    if led_select == 2:
        np[0] = [0, 0, 0]
        np[1] = [0, 0, 0]
        np[2] = [0, 0, 0]
        np[3] = [0, 0, 0]
        np.write()