/*
Copyright 2024.

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

package controller

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	tutorialv1 "github.com/yudar1024/operator-tutorial/api/v1"
)

// PrinterReconciler reconciles a Printer object
type PrinterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=tutorial.optutorial,resources=printers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=tutorial.optutorial,resources=printers/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=tutorial.optutorial,resources=printers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Printer object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/reconcile
func (r *PrinterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// TODO(user): your logic here
	var printer tutorialv1.Printer
	if err := r.Get(ctx, req.NamespacedName, &printer); err != nil {
		log.Error(err, "unable to fetch Printer")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}
	log.Info("Reconciling Printer", "Printer.Namespace", printer.Namespace, "Printer.Name", printer.Name)

	var configmap corev1.ConfigMap
	if err := r.Get(ctx, client.ObjectKey{Namespace: printer.Namespace, Name: printer.Name}, &configmap); err != nil {
		log.Info("Configmap not found, creating a new one")
		log.Info("Creating a new configmap")
		new_configmap := &corev1.ConfigMap{
			ObjectMeta: ctrl.ObjectMeta{Namespace: printer.Namespace, Name: printer.Name},
			Data: map[string]string{
				"configmapmsg": "Hello, Kubernetes configmap!",
				"mes2":"mes	2",
			},
		}
		if err := ctrl.SetControllerReference(&printer, new_configmap, r.Scheme); err != nil {
			log.Error(err, "unable to set owner reference")
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, new_configmap); err != nil {
			log.Error(err, "unable to create configmap")
			return ctrl.Result{}, err
		}
	} else {
		log.Info("Configmap found, skipping creation")
		// TODO(user): update configmap if needed
		return ctrl.Result{}, nil
	}
	

	var deployment appsv1.Deployment
	if err := r.Get(ctx, client.ObjectKey{Namespace: printer.Namespace, Name: printer.Name}, &deployment); err != nil {
		log.Info("Deployment not found, creating a new one")
		
		log.Info("Creating a new Deployment")
		var replicas int32 = 1
		new_deployment := &appsv1.Deployment{
			ObjectMeta: ctrl.ObjectMeta{Namespace: printer.Namespace, Name: printer.Name},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": printer.Name}},
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"app": printer.Name}},
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{
								Name:  "hello-operator",
								Image: "busybox",
								Command: []string{
									"sh",
									"-c",
									"echo Hello, Kubernetes!",
								},
							},
						},
					},
				},
			},
		}
		if err := ctrl.SetControllerReference(&printer, new_deployment, r.Scheme); err != nil {
			log.Error(err, "unable to set owner reference")
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, new_deployment); err != nil {
			log.Error(err, "unable to create Deployment")
			return ctrl.Result{}, err
		}
	} else {
		log.Info("Deployment found, skipping creation")
		// TODO(user): update deployment if needed

		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PrinterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tutorialv1.Printer{}).
		Complete(r)
}
