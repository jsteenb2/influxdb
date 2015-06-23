package graphite

import (
	"fmt"
	"strings"

	"github.com/influxdb/influxdb/toml"
	"github.com/influxdb/influxdb/tsdb"
)

const (
	// DefaultBindAddress is the default binding interface if none is specified.
	DefaultBindAddress = ":2003"

	// DefaultDatabase is the default database if none is specified.
	DefaultDatabase = "graphite"

	// DefaultProtocol is the default IP protocol used by the Graphite input.
	DefaultProtocol = "tcp"

	// DefaultConsistencyLevel is the default write consistency for the Graphite input.
	DefaultConsistencyLevel = "one"
)

// Config represents the configuration for Graphite endpoints.
type Config struct {
	BindAddress      string        `toml:"bind-address"`
	Database         string        `toml:"database"`
	Enabled          bool          `toml:"enabled"`
	Protocol         string        `toml:"protocol"`
	BatchSize        int           `toml:"batch-size"`
	BatchTimeout     toml.Duration `toml:"batch-timeout"`
	ConsistencyLevel string        `toml:"consistency-level"`
	Templates        []string      `toml:"templates"`
	Tags             []string      `toml:"tags"`
}

// NewConfig returns a new Config with defaults.
func NewConfig() Config {
	return Config{
		BindAddress:      DefaultBindAddress,
		Database:         DefaultDatabase,
		Protocol:         DefaultProtocol,
		ConsistencyLevel: DefaultConsistencyLevel,
	}
}

// WithDefaults takes the given config and returns a new config with any required
// default values set.
func (c *Config) WithDefaults() *Config {
	d := *c
	if d.BindAddress == "" {
		d.BindAddress = DefaultBindAddress
	}
	if d.Database == "" {
		d.Database = DefaultDatabase
	}
	if d.Protocol == "" {
		d.Protocol = DefaultProtocol
	}
	if d.ConsistencyLevel == "" {
		d.ConsistencyLevel = DefaultConsistencyLevel
	}
	return &d
}

func (c *Config) DefaultTags() tsdb.Tags {
	tags := tsdb.Tags{}
	for _, t := range c.Tags {
		parts := strings.Split(t, "=")
		tags[parts[0]] = parts[1]
	}
	return tags
}

func (c *Config) Validate() error {
	if err := c.validateTemplates(); err != nil {
		return err
	}

	if err := c.validateTags(); err != nil {
		return err
	}

	return nil
}

func (c *Config) validateTemplates() error {
	for i, t := range c.Templates {
		parts := strings.Fields(t)
		// Ensure template string is non-empty
		if len(parts) == 0 {
			return fmt.Errorf("missing template at position: %d", i)
		}
		if len(parts) == 1 && parts[0] == "" {
			return fmt.Errorf("missing template at position: %d", i)
		}

		if len(parts) > 3 {
			return fmt.Errorf("invalid template format: '%s'", t)
		}

		template := t
		filter := ""
		tags := ""
		if len(parts) >= 2 {
			filter = parts[0]
			template = parts[1]
		}

		if len(parts) == 3 {
			tags = parts[2]
		}

		// Validate the template has one and only one measurement
		if err := c.validateTemplate(template); err != nil {
			return err
		}
		// Validate filter expression is valid
		if err := c.validateFilter(filter); err != nil {
			return err
		}

		if tags != "" {
			// Validate tags
			for _, tagStr := range strings.Split(tags, ",") {
				if err := c.validateTag(tagStr); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (c *Config) validateTags() error {
	for _, t := range c.Tags {
		if err := c.validateTag(t); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) validateTemplate(template string) error {
	hasMeasurement := false
	for _, p := range strings.Split(template, ".") {
		if p == "measurement" || p == "measurement*" {
			if hasMeasurement {
				return fmt.Errorf("multiple measurements in template `%s`", template)
			}
			hasMeasurement = true
		}
	}

	if !hasMeasurement {
		return fmt.Errorf("no measurement in template `%s`", template)
	}

	return nil
}

func (c *Config) validateFilter(filter string) error {
	for _, p := range strings.Split(filter, ".") {
		if p == "" {
			return fmt.Errorf("filter contains blank section: %s", filter)
		}

		if strings.Contains(p, "*") && p != "*" {
			return fmt.Errorf("invalid filter wildcard section: %s", filter)
		}
	}
	return nil
}

func (c *Config) validateTag(keyValue string) error {
	parts := strings.Split(keyValue, "=")
	if len(parts) != 2 {
		return fmt.Errorf("invalid template tags: '%s'", keyValue)
	}

	if parts[0] == "" || parts[1] == "" {
		return fmt.Errorf("invalid template tags: %s'", keyValue)
	}

	return nil
}
