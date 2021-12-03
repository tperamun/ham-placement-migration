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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1alpha1 "github.com/ham-placement/api/v1alpha1"
)

var PlacementDecisionMaker DecisionMaker

// blank assignment to verify that ReconcilePlacementRule implements reconcile.Reconciler
var _ reconcile.Reconciler = &PlacementRuleReconciler{}

// PlacementRuleReconciler reconciles a PlacementRule object
type PlacementRuleReconciler struct {
	client        client.Client
	scheme        *runtime.Scheme
	dynamicClient dynamic.Interface
	decisionMaker DecisionMaker
}

// +kubebuilder:rbac:groups=core.hybridapp.io,resources=placementrules,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.hybridapp.io,resources=placementrules/status,verbs=get;update;patch

func (r *PlacementRuleReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()

	klog.Info("Reconciling PlacementRule ", req.NamespacedName)

	// Fetch the PlacementRule instance
	instance := &corev1alpha1.PlacementRule{}

	err := r.client.Get(context.TODO(), req.NamespacedName, instance)

	if err != nil {
		if errors.IsNotFound(err) {

			return ctrl.Result{}, nil
		}

		// Error reading the object - requeue the request
		return ctrl.Result{}, err
	}

	// Step 1: generate new candidates from spec
	ncans, err := r.generateCandidates(instance)
	if err != nil {
		klog.Error("Failed to generate candidates for decision with error: ", err)
	}

	// if spec has been changed, reset it
	if instance.Status.ObservedGeneration != instance.GetGeneration() || !isSameCandidateList(ncans, instance) {
		err = r.resetDecisionMakingProcess(ncans, instance)
		if err != nil {
			klog.Error("Following error occurred during resetDecisionMakingProcess: ", err)
		}

		return ctrl.Result{}, err
	}

	return ctrl.Result{}, r.continueDecisionMakingProcess(instance)
}
func (r *PlacementRuleReconciler) resetDecisionMakingProcess(candidates []corev1.ObjectReference, instance *corev1alpha1.PlacementRule) error {
	instance.Status.ObservedGeneration = instance.GetGeneration()
	now := metav1.Now()
	instance.Status.LastUpdateTime = &now
	instance.Status.Candidates = candidates
	instance.Status.Recommendations = nil
	instance.Status.Eliminators = nil

	r.decisionMaker.ResetDecisionMakingProcess(candidates, instance)

	return r.client.Status().Update(context.TODO(), instance)
}

func (r *PlacementRuleReconciler) continueDecisionMakingProcess(instance *corev1alpha1.PlacementRule) error {
	readytodecide := true

	for _, adv := range instance.Spec.Advisors {
		if instance.Status.Recommendations == nil {
			readytodecide = false
			break
		}

		if _, ok := instance.Status.Recommendations[adv.Name]; !ok {
			readytodecide = false
			break
		}
	}

	if readytodecide && r.decisionMaker.ContinueDecisionMakingProcess(instance) {
		return r.client.Status().Update(context.TODO(), instance)
	}

	return nil
}

func (r *PlacementRuleReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// Create a new controller
	// r := newReconciler(mgr)
	// // c, err := controller.New("placementrule-controller", mgr, controller.Options{Reconciler: r})

	// return ctrl.NewControllerManagedBy(mgr).
	// 	For(&corev1alpha1.PlacementRule{}).
	// 	Owns(&source.Kind{Type: &corev1alpha1.PlacementRule{}}, &libhandler.InstrumentedEnqueueRequestForObject{})
	// Complete(r)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	if PlacementDecisionMaker == nil {
		PlacementDecisionMaker = &DefaultDecisionMaker{}
	}

	rec := &PlacementRuleReconciler{
		client:        mgr.GetClient(),
		scheme:        mgr.GetScheme(),
		dynamicClient: dynamic.NewForConfigOrDie(mgr.GetConfig()),
		decisionMaker: PlacementDecisionMaker,
	}

	return rec
}
