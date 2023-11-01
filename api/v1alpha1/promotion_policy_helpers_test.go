package v1alpha1

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestGetPromotionPolicy(t *testing.T) {
	scheme := k8sruntime.NewScheme()
	require.NoError(t, SchemeBuilder.AddToScheme(scheme))

	testCases := []struct {
		name       string
		client     client.Client
		assertions func(*PromotionPolicy, error)
	}{
		{
			name:   "not found",
			client: fake.NewClientBuilder().WithScheme(scheme).Build(),
			assertions: func(policy *PromotionPolicy, err error) {
				require.NoError(t, err)
				require.Nil(t, policy)
			},
		},

		{
			name: "found",
			client: fake.NewClientBuilder().WithScheme(scheme).WithObjects(
				&PromotionPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "fake-promotion-policy",
						Namespace: "fake-namespace",
					},
				},
			).Build(),
			assertions: func(policy *PromotionPolicy, err error) {
				require.NoError(t, err)
				require.Equal(t, "fake-promotion-policy", policy.Name)
				require.Equal(t, "fake-namespace", policy.Namespace)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			policy, err := GetPromotionPolicy(
				context.Background(),
				testCase.client,
				types.NamespacedName{
					Namespace: "fake-namespace",
					Name:      "fake-promotion-policy",
				},
			)
			testCase.assertions(policy, err)
		})
	}
}