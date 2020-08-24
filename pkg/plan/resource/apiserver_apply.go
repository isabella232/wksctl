package resource

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/cluster-api-provider-existinginfra/pkg/plan"
	"github.com/weaveworks/cluster-api-provider-existinginfra/pkg/plan/resource"
	"github.com/weaveworks/cluster-api-provider-existinginfra/pkg/utilities/manifest"
	"github.com/weaveworks/libgitops/pkg/serializer"
)

// ApiserverApply is a resource applying the provided manifest.
// It doesn't realise any state, Apply will always apply the manifest.
type ApiserverApply struct {
	resource.Base

	// Manifest is the actual YAML/JSON content of the manifest to apply.
	// If this is provided, then there is no need to provide ManifestPath, but
	// Filename should be provided in order to name the remote manifest file.
	Manifest []byte `structs:"manifest"`
	// ManifestPath is the path to the manifest to apply.
	// If this is provided, then there is no need to provide Manifest.
	ManifestPath fmt.Stringer `structs:"manifestPath"`
	// ManifestURL is the URL of a remote manifest; if specified,
	// neither Filename, Manifest, nor ManifestPath should be specified.
	ManifestURL fmt.Stringer `structs:"manifestURL"`
	// WaitCondition, if not empty, makes Apply() perform "apiserver wait --for=<value>" on the resource.
	Namespace fmt.Stringer `structs:"namespace"`
	// OpaqueManifest is an alternative to Manifest for a resource to
	// apply whose content should not be exposed in a serialized plan.
	// If this is provided, then there is no need to provide
	// ManifestPath, but Filename should be provided in order to name
	// the remote manifest file.
	OpaqueManifest []byte `structs:"-" plan:"hide"`
	// ManifestPath is the path to the manifest to apply.
	// If this is provided, then there is no need to provide Manifest.
	// For example, waiting for "condition=established" is required after creating a CRD - see issue #530.
	WaitCondition string `structs:"afterApplyWaitsFor"`
}

var _ plan.Resource = plan.RegisterResource(&ApiserverApply{})

// State implements plan.Resource.
func (a *ApiserverApply) State() plan.State {
	return resource.ToState(a)
}

func (a *ApiserverApply) content() ([]byte, error) {
	if a.Manifest != nil {
		return a.Manifest, nil
	}

	if a.OpaqueManifest != nil {
		return a.OpaqueManifest, nil
	}

	if url := str(a.ManifestURL); url != "" {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		return ioutil.ReadAll(resp.Body)
	}

	if path := str(a.ManifestPath); path != "" {
		return ioutil.ReadFile(path)
	}

	return nil, errors.New("no content provided")
}

// Apply performs a "apiserver apply" as specified in the receiver.
func (a *ApiserverApply) Apply(runner plan.Runner, diff plan.Diff) (bool, error) {

	// Get the manifest content.
	c, err := a.content()
	if err != nil {
		return false, err
	}

	if str(a.Namespace) != "" {
		content, err := manifest.WithNamespace(serializer.FromBytes(c), str(a.Namespace))
		if err != nil {
			return false, err
		}
		if len(content) != 0 {
			c = content
		}
	}

	if err := apiserverApply(runner, apiserverApplyArgs{
		Content:       c,
		WaitCondition: a.WaitCondition,
	}); err != nil {
		return false, err
	}

	return true, nil
}

type apiserverApplyArgs struct {
	// Content is the YAML manifest to be applied. Must be non-empty.
	Content []byte
	// WaitCondition, if non-empty, makes apiserverApply do "apiserver wait --for=<value>" on the applied resource.
	WaitCondition string
}

func apiserverApply(r plan.Runner, args apiserverApplyArgs) error {
	// Run apiserver apply.
	// if err := apiserverRemoteApply(path, r); err != nil {
	// 	return errors.Wrap(err, "apiserver apply")
	// }

	// // Run apiserver wait, if requested.
	// if args.WaitCondition != "" {
	// 	cmd := fmt.Sprintf("kubectl wait --for=%q -f %q", args.WaitCondition, path)
	// 	if _, err := r.RunCommand(withoutProxy(cmd), nil); err != nil {
	// 		return errors.Wrap(err, "kubectl wait")
	// 	}
	// }

	// Great success!
	return nil
}

func apiserverRemoteApply(remoteURL string, runner plan.Runner) error {
	log.Debug("applying")
	//cmd := fmt.Sprintf("apiserver apply -f %q", remoteURL)

	// if stdouterr, err := runner.RunCommand(withoutProxy(cmd), nil); err != nil {
	// 	log.WithField("stdouterr", stdouterr).WithField("URL", remoteURL).Debug("failed to apply Kubernetes manifest")
	// 	return errors.Wrapf(err, "failed to apply manifest %s; output %s", remoteURL, stdouterr)
	// }
	return nil
}
