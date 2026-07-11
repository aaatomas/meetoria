package service

var AllowedColors = []string{
	"red", "pink", "purple", "indigo", "blue", "teal", "green", "lime", "amber", "orange", "grey",
}

var allowedColorSet = func() map[string]struct{} {
	set := make(map[string]struct{}, len(AllowedColors))
	for _, color := range AllowedColors {
		set[color] = struct{}{}
	}
	return set
}()

func NormalizeColor(color string) string {
	if _, ok := allowedColorSet[color]; ok {
		return color
	}
	return "teal"
}
