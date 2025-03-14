package main

import (
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/pointer"

	kargoapi "github.com/akuity/kargo/api/v1alpha1"
	typesv1alpha1 "github.com/akuity/kargo/internal/api/types/v1alpha1"
	"github.com/akuity/kargo/internal/cli/client"
	"github.com/akuity/kargo/internal/cli/option"
	versionpkg "github.com/akuity/kargo/internal/version"
	svcv1alpha1 "github.com/akuity/kargo/pkg/api/service/v1alpha1"
)

func newVersionCommand(opt *option.Option) *cobra.Command {
	cmd := &cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			var serverVersion *svcv1alpha1.VersionInfo
			if !opt.UseLocalServer {
				kargoCli, err := client.GetClientFromConfig(ctx, opt)
				if err != nil {
					if !client.IsConfigNotFoundErr(err) {
						return errors.Wrap(err, "get Kargo client from config")
					}
					// Skip initializing server version if config not found (not logged in).
				} else {
					resp, err := kargoCli.GetVersionInfo(ctx, connect.NewRequest(&svcv1alpha1.GetVersionInfoRequest{}))
					if err != nil {
						return errors.Wrap(err, "get version info from server")
					}
					serverVersion = resp.Msg.GetVersionInfo()
				}
			}
			cliVersion := typesv1alpha1.ToVersionProto(versionpkg.GetVersion())

			if pointer.StringDeref(opt.PrintFlags.OutputFormat, "") == "" {
				fmt.Println("Client Version:", cliVersion.GetVersion())
				if serverVersion != nil {
					fmt.Println("Server Version:", serverVersion.GetVersion())
				}
				return nil
			}

			printer, err := opt.PrintFlags.ToPrinter()
			if err != nil {
				return errors.Wrap(err, "new printer")
			}
			obj, err := componentVersionsToRuntimeObject(&svcv1alpha1.ComponentVersions{
				Server: serverVersion,
				Cli:    cliVersion,
			})
			if err != nil {
				return errors.Wrap(err, "map component versions to runtime object")
			}
			return printer.PrintObj(obj, opt.IOStreams.Out)
		},
	}
	opt.PrintFlags.AddFlags(cmd)
	return cmd
}

func componentVersionsToRuntimeObject(v *svcv1alpha1.ComponentVersions) (runtime.Object, error) {
	data, err := protojson.Marshal(v)
	if err != nil {
		return nil, errors.Wrap(err, "marshal component versions")
	}
	var content map[string]any
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, errors.Wrap(err, "unmarshal component versions")
	}
	u := &unstructured.Unstructured{}
	u.SetUnstructuredContent(content)
	u.SetAPIVersion(kargoapi.GroupVersion.String())
	u.SetKind("ComponentVersions")
	return u, nil
}
