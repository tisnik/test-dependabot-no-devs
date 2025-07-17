(*
    (C) Copyright 2023  Pavel Tisnovsky

    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at

        http://www.apache.org/licenses/LICENSE-2.0

    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*)



let scale_component x =
  int_of_float (255.*.x)
;;


let scale_rgb r g b =
  (scale_component r,
   scale_component g,
   scale_component b)
;;


let hsv_to_rgb_ h s v =
  let h = 
    match h with
    | 1.0 -> 0.0
    | _ -> h
  in
  let i = int_of_float (h*.6.0) in
  let f = h *. 6.0 -. (float i) in
  let w = v *. (1.0 -. s) in
  let q = v *. (1.0 -. s*.f) in
  let t = v *. (1.0 -. s*.(1.0 -. f)) in
  match i with
  | 0 -> scale_rgb v t w
  | 1 -> scale_rgb q v w
  | 2 -> scale_rgb w v t
  | 3 -> scale_rgb w q v
  | 4 -> scale_rgb t w v
  | 5 -> scale_rgb v w q
  | _ -> (0, 0, 0)
;;


let hsv_to_rgb h s v =
  match s with
  | 0.0 -> (scale_rgb v v v)
  | _ -> (hsv_to_rgb_ h s v)
;;


type basic_color =
  | Black
  | Red
  | Green
  | Yellow
  | Blue
  | Magenta
  | Cyan
  | White
;;


type brightness =
  | Dark
  | Bright
;;


type color =
  | BasicColor of basic_color * brightness
  | Gray of int
  | RGB of int * int * int
  | HSV of float * float * float
  | Mix of float * color * color
;;


let basic_color_to_rgb = function
  | Black -> (0, 0, 0)
  | Red -> (255, 0,0)
  | Green -> (0, 255, 0)
  | Yellow -> (255, 255, 0)
  | Blue -> (0, 0, 255)
  | Magenta -> (255, 0, 255)
  | Cyan -> (0, 255, 255)
  | White -> (255, 255, 255)
;;


let darker = function
  | (r, g, b) -> (r/2, g/2, b/2)
;;


let brightness rgb brightess =
  match brightess with
  | Dark -> darker rgb
  | Bright -> rgb
;;


let mix_components ratio c1 c2 =
  int_of_float ((float c1) *. ratio +. (float c2) *. (1.0 -. ratio))
;;


let rec mix_colors ratio color1 color2 =
  let (r1, g1, b1) = to_rgb color1 in
  let (r2, g2, b2) = to_rgb color2 in
  (mix_components ratio r1 r2, mix_components ratio g1 g2, mix_components ratio b1 b2)
and
  to_rgb = function
  | BasicColor (c, b) -> brightness (basic_color_to_rgb c) b
  | Gray g -> (g, g, g)
  | RGB(r,g,b) -> (r, g, b)
  | HSV(h,s,v) -> hsv_to_rgb h s v
  | Mix(ratio, color1, color2) -> mix_colors ratio color1 color2
;;


let c1 = BasicColor(Black, Dark);;
to_rgb c1;;

let c2 = BasicColor(Black, Bright);;
to_rgb c2;;

let c3 = BasicColor(Red, Dark);;
to_rgb c3;;

let c4 = BasicColor(Red, Bright);;
to_rgb c4;;

let g1 = Gray(0);;
to_rgb g1;;

let g2 = Gray(255);;
to_rgb g2;;

let rgb1 = RGB(0, 10, 20);;
to_rgb rgb1;;

let rgb2 = RGB(0, 0, 255);;
to_rgb rgb2;;

let rgb3 = RGB(255, 255, 255);;
to_rgb rgb3;;

let hsv1 = HSV(0.0, 0.0, 1.0);;
to_rgb hsv1;;

let hsv2 = HSV(0.0, 0.0, 0.5);;
to_rgb hsv2;;

let hsv3 = HSV(0.0, 1.0, 1.0);;
to_rgb hsv3;;

let hsv4 = HSV(0.3333, 1.0, 1.0);;
to_rgb hsv4;;

let hsv5 = HSV(0.6666, 1.0, 1.0);;
to_rgb hsv5;;

let hsv6 = HSV(1.0, 1.0, 1.0);;
to_rgb hsv6;;

let hsv7 = HSV(1.0, 0.5, 0.5);;
to_rgb hsv7;; 

let mixed1 = Mix(0.0, BasicColor(Red, Bright), BasicColor(Blue, Bright));;
to_rgb mixed1;;

let mixed2 = Mix(0.5, BasicColor(Red, Bright), BasicColor(Blue, Bright));;
to_rgb mixed2;;

let mixed3 = Mix(1.0, BasicColor(Red, Bright), BasicColor(Blue, Bright));;
to_rgb mixed3;;
