package state

import (
	"context"
	"testing"

	serverlessv1alpha2 "github.com/kyma-project/serverless/api/v1alpha2"
	"github.com/kyma-project/serverless/internal/controller/fsm"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_sFnConfigurationReady(t *testing.T) {
	t.Run("should set condition and go to the next state", func(t *testing.T) {
		// Arrange
		// machine with our function
		m := fsm.StateMachine{}

		// Act
		next, result, err := sFnConfigurationReady(context.Background(), &m)

		// Assert
		// no errors
		require.Nil(t, err)
		// without stopping processing
		require.Nil(t, result)
		// with expected next state
		require.NotNil(t, next)
		requireEqualFunc(t, sFnHandleDeployment, next)
		// function has proper condition
		requireContainsCondition(t, m.State.Function.Status,
			serverlessv1alpha2.ConditionConfigurationReady,
			metav1.ConditionTrue,
			serverlessv1alpha2.ConditionReasonFunctionSpecValidated,
			"Function configured")
	})
}
