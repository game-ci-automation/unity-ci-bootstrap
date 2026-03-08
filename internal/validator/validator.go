package validator

import "regexp"

var unityVersionRegex = regexp.MustCompile(`^\d+\.\d+\.\d+f\d+$`)

var supportedPlatforms = map[string]bool{
	"WebGL":               true,
	"Android":             true,
	"iOS":                 true,
	"StandaloneLinux64":   true,
	"StandaloneWindows64": true,
}

func ValidateUnityVersion(version string) bool {
	return unityVersionRegex.MatchString(version)
}

func ValidatePlatform(platform string) bool {
	return supportedPlatforms[platform]
}
