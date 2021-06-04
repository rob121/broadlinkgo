package main


var icons map[string]string

type Pair struct {
	Key   string
	Value string
}

type PairList []Pair


func (p PairList) Len() int           { return len(p) }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }

func init() {

	icons = map[string]string{
		"fa-adjust": "Adjust",
		"fa-apple": "Apple",
		"fa-bolt": "Power Alt",
		"fa-play": "Play",
		"fa-border-all": "Split Screen",
		"fa-square": "Single Screen",
		"fa-pause": "Pause",
		"fa-stop": "Stop",
		"fa-circle": "Record",
		"fa-info-circle": "Info",
		"fa-check-circle": "Check",
		"fa-power-off": "Power",
		"fa-times-circle": "Exit",
		"fa-bars": "Menu",
		"fa-undo-alt": "Return",
		"fa-forward": "Forward",
		"fa-fast-forward": "Fast Forward",
		"fa-backward": "Backward",
		"fa-fast-backward": "Fast Backward",
		"fa-volume-up": "Volume Up",
		"fa-volume-down": "Volume Down",
		"fa-volume-mute": "Mute",
		"fa-plus": "Plus",
		"fa-minus": "Minus",
		"fa-plus-square": "Plus",
		"fa-minus-square": "Minus",
		"fa-num-1": "1 Button",
		"fa-num-2": "2 Button",
		"fa-num-3": "3 Button",
		"fa-num-4": "4 Button",
		"fa-num-5": "5 Button",
		"fa-num-6": "6 Button",
		"fa-num-7": "7 Button",
		"fa-num-8": "8 Button",
		"fa-num-9": "9 Button",
		"fa-num-0": "0 Button",
		"fa-remote-menu": "Menu",
		"fa-remote-exit": "Exit",
		"fa-tv": "TV",
		"fa-laptop": "Computer",
		"fa-server": "Component",
		"fa-hdd": "Device",
		"fa-tools": "Tools",
		"fa-caret-up": "Up",
		"fa-caret-down": "Down",
		"fa-caret-left": "Left",
		"fa-caret-right": "Right",
		"fa-music": "Music",
		"fa-film": "Movie",
		"fab fa-youtube": "Youtube",
	}



}