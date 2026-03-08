package docker

import (
	"fmt"
)

var platformTags = map[string]string{
	"WebGL":               "webgl",
	"Android":             "android",
	"iOS":                 "ios",
	"StandaloneLinux64":   "linux-il2cpp",
	"StandaloneWindows64": "windows-mono",
}

func ResolveImageTag(version, platform string) (string, error) {
	tag, ok := platformTags[platform]
	if !ok {
		return "", fmt.Errorf("unsupported platform: %q", platform)
	}
	return fmt.Sprintf("unityci/editor:ubuntu-%s-%s-3", version, tag), nil
}
