// Package config loads gibson's settings with viper. Every setting has a
// key, and its value comes from, in order of precedence, a flag, a GIBSON_
// environment variable, the config file, or the flag's default.
//
// Keys are dotted, and their environment variable is the key in upper case
// with "." and "-" turned into "_", prefixed with GIBSON_: ssh.host-key is
// GIBSON_SSH_HOST_KEY.
package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/log/v2"
	"github.com/spf13/cast"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/spechtlabs/gibson/internal/scene"
)

// Look keys, shared by every command.
const (
	Theme    = "theme"
	Specimen = "specimen"
	Vessel   = "vessel"
	Tagline  = "tagline"
	FPS      = "fps"
	Start    = "start"
	Pausable = "pausable"
	Music    = "music"
)

// SSH server keys; healthcheck reads Listen too.
const (
	Listen      = "ssh.listen"
	HostKey     = "ssh.host-key"
	MaxConns    = "ssh.max-conns"
	MaxSessions = "ssh.max-sessions"
	Rate        = "ssh.rate"
	Burst       = "ssh.burst"
	MaxDuration = "ssh.max-duration"
	ProxyProto  = "ssh.proxy-protocol"
)

// Offline render keys.
const (
	RenderPNG  = "render.png"
	RenderGIF  = "render.gif"
	RenderAt   = "render.at"
	RenderDur  = "render.dur"
	RenderSize = "render.size"
)

// New returns a viper that reads GIBSON_ environment variables.
func New() *viper.Viper {
	v := viper.New()
	v.SetEnvPrefix("GIBSON")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
	return v
}

// Load reads the config file into v: path when it's set, which must exist,
// and otherwise the first gibson.yaml in ., $XDG_CONFIG_HOME/gibson and
// /etc/gibson, if there is one.
func Load(v *viper.Viper, path string) error {
	if path != "" {
		v.SetConfigFile(path)
	} else {
		// No SetConfigType: with one, viper would also try a bare "gibson",
		// which in the working directory is likely the binary.
		v.SetConfigName("gibson")
		v.AddConfigPath(".")
		if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
			v.AddConfigPath(filepath.Join(xdg, "gibson"))
		} else if home, err := os.UserHomeDir(); err == nil {
			v.AddConfigPath(filepath.Join(home, ".config", "gibson"))
		}
		v.AddConfigPath("/etc/gibson")
	}
	if err := v.ReadInConfig(); err != nil {
		if errors.As(err, new(viper.ConfigFileNotFoundError)) {
			return nil
		}
		return fmt.Errorf("read config: %w", err)
	}
	// Debug, since the container healthcheck loads it every 30s; serve logs
	// the file it runs with at info.
	log.Debug("loaded config", "file", v.ConfigFileUsed())
	return nil
}

// Bind binds the flags in fs to the keys they're mapped to. Commands bind
// when they run, since several bind a flag of the same name to one key
// (--pausable, --listen) and viper keeps only the last binding.
//
// Bind also checks that each key's value, wherever it came from, parses as
// its flag's type: viper's GetInt and friends turn a value they can't parse,
// say GIBSON_SSH_MAX_DURATION=soon, into 0, which would quietly lift a limit.
func Bind(v *viper.Viper, fs *pflag.FlagSet, keys map[string]string) error {
	for name, key := range keys {
		f := fs.Lookup(name)
		if err := v.BindPFlag(key, f); err != nil {
			return fmt.Errorf("bind --%s: %w", name, err)
		}
		if err := check(f.Value.Type(), v.Get(key)); err != nil {
			return fmt.Errorf("invalid %s: %w", key, err)
		}
	}
	return nil
}

func check(typ string, val any) error {
	var err error
	switch typ {
	case "int":
		_, err = cast.ToIntE(val)
	case "float64":
		_, err = cast.ToFloat64E(val)
	case "bool":
		_, err = cast.ToBoolE(val)
	case "duration":
		_, err = cast.ToDurationE(val)
	}
	return err
}

// Look returns the look the look keys pick, and an error naming the valid
// values for any it doesn't know.
func Look(v *viper.Viper) (scene.Look, error) {
	l := scene.Look{
		Theme:    v.GetString(Theme),
		Specimen: v.GetString(Specimen),
		Vessel:   v.GetString(Vessel),
		Tagline:  v.GetString(Tagline),
	}
	if _, err := l.Options(); err != nil {
		return scene.Look{}, err
	}
	if fps := v.GetInt(FPS); fps <= 0 {
		return scene.Look{}, fmt.Errorf("fps must be positive, got %d", fps)
	}
	return l, nil
}
