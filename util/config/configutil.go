package config

import (
	"strings"

	"github.com/spf13/viper"
)

func GetOrganizationUrl() string {
	return GetStringValue("org_url")
}

func GetOrganizationName() string {
	parts := strings.Split(GetStringValue("org_url"), "/")
	return parts[len(parts)-1]
}

func SetOrganizationUrl(url string) {
	viper.Set("org_url", url)
	viper.WriteConfig()
}

func GetPat() string {
	return GetStringValue("pat")
}

func SetPat(pat string) {
	viper.Set("pat", pat)
	viper.WriteConfig()
}

func GetStringValue(key string) string {
	value := viper.Get(key)
	if value == nil {
		return ""
	}
	return value.(string)
}
