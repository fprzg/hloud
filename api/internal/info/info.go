package info

import "fmt"

type Build struct {
	Version string
	Time    string
}

type Config struct {
	Port       int
	Env        string
	StorageDir string
}

func (c *Config) GetPort() string {
	// NOTE(Farid): Check if we benefit from cacheing this string so it doesn't get allocated every time we call the function
	return fmt.Sprintf(":%d", c.Port)
}
