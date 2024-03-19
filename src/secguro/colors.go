package main

var Black = "0;30"
var DarkGray = "1;30"
var Red = "0;31"
var LightRed = "1;31"
var Green = "0;32"
var LightGreen = "1;32"
var Brown = "0;33"
var Yellow = "1;33"
var Blue = "0;34"
var LightBlue = "1;34"
var Purple = "0;35"
var LightPurple = "1;35"
var Cyan = "0;36"
var LightCyan = "1;36"
var LightGray = "0;37"
var White = "1;37"
var NoColor = "0"

func changeColor(color string) string {
	return "\033[" + color + "m"
}
