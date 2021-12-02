/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1alpha1 "github.com/ham-placement/api/v1alpha1"
)

var PlacementDecisionMaker DecisionMaker

// PlacementRuleReconciler reconciles a PlacementRule object
type PlacementRuleReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme        *runtime.Scheme
	dynamicClient dynamic.Interface
	decisionMaker DecisionMaker
}

// +kubebuilder:rbac:groups=core.hybridapp.io,resources=placementrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.hybridapp.io,resources=placementrules/status,verbs=get;update;patch

func (r *PlacementRuleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	_ = r.Log.WithValues("placementrule", req.NamespacedName)

	return ctrl.Result{}, nil
}

func (r *PlacementRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// Create a new controller

	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.PlacementRule{}).
		Complete(r)
}
