package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/realbucksavage/openrgb-go"
	"github.com/sirupsen/logrus"
)

const stateFile = "~/.config/OpenRGB-switch/state"
const colorFile = "~/.config/OpenRGB-switch/colors"

type Color struct {
	R uint8
	G uint8
	B uint8
}

type Controller struct {
	Colors  []Color
	Current uint
}

func (ctrl *Controller) SetState(val uint) error {
	err := os.WriteFile(stateFile, []byte(fmt.Sprint(val)), 0644)
	if err != nil {
		ErrorLog(err, "error writing state file")
		return err
	}
	ctrl.Current = val
	return nil
}
func (ctrl *Controller) GetNext() (uint, Color) {
	nxt := ctrl.Current + 1
	if int(nxt) >= len(ctrl.Colors) {
		nxt = 0
	}
	return nxt, ctrl.Colors[nxt]
}
func (ctrl *Controller) SetColor(val Color) error {
	conn, err := openrgb.Connect("localhost", 6742)
	if err != nil {
		ErrorLog(err, "error connecting to OpenRGB server")
		return err
	}
	defer conn.Close()
	conCount, err := conn.GetControllerCount()
	if err != nil {
		ErrorLog(err, "error getting OpenRGB controllers")
		return err
	}
	for i := 0; i < conCount; i++ {
		controller, _ := conn.GetDeviceController(i)
		colors := make([]openrgb.Color, len(controller.Colors))
		for i := 0; i < len(colors); i++ {
			colors[i] = openrgb.Color{Red: uint8(val.R), Green: uint8(val.G), Blue: uint8(val.B)}
		}
		if err := conn.UpdateLEDs(i, colors); err != nil {
			ErrorLog(err, "error updating LED state")
		}
	}
	return nil
}
func NewController() (Controller, error) {
	ctrl := Controller{}
	err := parseColorFile(&ctrl)
	if err != nil {
		return ctrl, err
	}
	err = parseCurrent(&ctrl)
	if err != nil {
		return ctrl, err
	}
	return ctrl, nil
}
func parseColorFile(ctrl *Controller) error {
	dat, err := os.Open(colorFile)
	if os.IsNotExist(err) {
		if err = os.WriteFile(colorFile, []byte("255-0-0\n0-255-0\n0-0-255\n255-255-255"), 0644); err != nil {
			logrus.Error(err, "error writing default color file")
			return err
		}
		return parseColorFile(ctrl)
	} else if err != nil {
		ErrorLog(err, "error reading colors file")
		return err
	}
	defer dat.Close()
	colorFileScanner := bufio.NewScanner(dat)
	colorFileScanner.Split(bufio.ScanLines)
	for colorFileScanner.Scan() {
		val := colorFileScanner.Text()
		logrus.WithField("val", val).Debug("new color found in colors file")
		valSplit := strings.Split(val, "-")
		valSplitInt := make([]uint64, 3)
		for i, s := range valSplit {
			v, err := strconv.ParseUint(s, 10, 64)
			if err != nil {
				ErrorLog(err, "error parsing color")
				return err
			}
			valSplitInt[i] = v
		}
		clr := Color{
			R: uint8(valSplitInt[0]),
			G: uint8(valSplitInt[1]),
			B: uint8(valSplitInt[2]),
		}
		logrus.WithField("val", clr).Debug("new color parsed from colors file")
		ctrl.Colors = append(ctrl.Colors, clr)
	}
	return nil
}
func parseCurrent(ctrl *Controller) error {
	dat, err := os.ReadFile(stateFile)
	if os.IsNotExist(err) {
		logrus.Warn("state file not found")
		if err = ctrl.SetState(0); err != nil {
			return err
		}
		return nil
	} else if err != nil {
		ErrorLog(err, "error reading state file")
		return err
	}
	cu, err := strconv.ParseUint(string(dat), 10, 64)
	if err != nil {
		ErrorLog(err, "error parsing state")
		return err
	}
	ctrl.Current = uint(cu)
	logrus.WithField("val", ctrl.Current).Debug("current state parsed")
	return nil
}
