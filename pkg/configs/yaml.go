package configs

import (
	"fmt"

	"chainguard.dev/melange/pkg/build"
	"github.com/chainguard-dev/yam/pkg/yam/formatted"
	"github.com/dprotaso/go-yit"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

// A yamlUpdater is a function that mutates a YAML AST. The function is also
// given a build.Configuration in case implementations require it for context.
type yamlUpdater func(build.Configuration, *yaml.Node) error

func (i *Index) newYAMLUpdateFunc(updateYAML yamlUpdater) updateFunc {
	return func(e Entry) error {
		root := e.YAMLRoot()

		cfg := e.Configuration()
		if cfg == nil {
			return errors.New("nil configuration")
		}

		err := updateYAML(*cfg, root)
		if err != nil {
			return err
		}

		file, err := i.fsys.OpenAsWritable(e.Path())
		if err != nil {
			return fmt.Errorf("unable to update %q: %w", e.Path(), err)
		}
		defer file.Close()

		err = i.fsys.Truncate(e.Path(), 0)
		if err != nil {
			return fmt.Errorf("unable to update %q: %w", e.Path(), err)
		}

		encoder := formatted.NewEncoder(file).AutomaticConfig()

		err = encoder.Encode(root)
		if err != nil {
			return fmt.Errorf("unable to encode updated YAML: %w", err)
		}

		return nil
	}
}

func yamlNodeForKey(root *yaml.Node, key string) *yaml.Node {
	rootMap := root.Content[0]

	iter := yit.FromNode(rootMap).ValuesForMap(yit.WithValue(key), yit.All)
	advNode, ok := iter()
	if ok {
		return advNode
	}

	mapKey := &yaml.Node{Value: key, Tag: "!!str", Kind: yaml.ScalarNode}
	rootMap.Content = append(rootMap.Content, mapKey)
	mapValue := &yaml.Node{Tag: "!!map", Kind: yaml.MappingNode}
	rootMap.Content = append(rootMap.Content, mapValue)

	return mapValue
}
