/*
Copyright 2023.

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
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	hellov1 "github.com/KR411-prog/helloworld-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// HelloworldReconciler reconciles a Helloworld object
type HelloworldReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=hello.mydomain,resources=helloworlds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=hello.mydomain,resources=helloworlds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=hello.mydomain,resources=helloworlds/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Helloworld object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.4/pkg/reconcile
func (r *HelloworldReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	hellowordspec := &hellov1.Helloworld{}
	err := r.Client.Get(ctx, req.NamespacedName, hellowordspec)

	if err != nil {
		return ctrl.Result{}, err
	}

	// name of our custom finalizer
	myFinalizerName := "mydomain.helloworld/finalizer"

	// examine DeletionTimestamp to determine if object is under deletion
	if hellowordspec.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !controllerutil.ContainsFinalizer(hellowordspec, myFinalizerName) {
			controllerutil.AddFinalizer(hellowordspec, myFinalizerName)
			if err := r.Update(ctx, hellowordspec); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if controllerutil.ContainsFinalizer(hellowordspec, myFinalizerName) {
			// our finalizer is present, so lets handle any external dependency
			//   if err := r.deleteExternalResources(hellowordspec); err != nil {
			// 	  // if fail to delete the external dependency here, return with error
			// 	  // so that it can be retried
			// 	  return ctrl.Result{}, err
			//   }

			if err := r.deleteExternalResources(ctx, hellowordspec, req.NamespacedName.Namespace); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return ctrl.Result{}, err
			}
			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(hellowordspec, myFinalizerName)
			if err := r.Update(ctx, hellowordspec); err != nil {
				return ctrl.Result{}, err
			}
		}
		// Stop reconciliation as the item is being deleted
		return ctrl.Result{}, nil
	}

	fmt.Println("hellowordspec is ................")
	fmt.Println(hellowordspec)
	fmt.Println("req.NamespacedName is ................")
	fmt.Println(req.NamespacedName)
	// TODO(user): your logic here
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: hellowordspec.Spec.PodName, Namespace: req.NamespacedName.Namespace},
		//TypeMeta: metav1.TypeMeta{Kind: ,APIVersion:},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: hellowordspec.Spec.PodName,
					Image: "nginx"},
			},
		},
	}
	fmt.Printf("Creating pod with name %s\n", hellowordspec.Spec.PodName)
	r.Client.Create(ctx, &pod)
	return ctrl.Result{}, nil
}

func (r *HelloworldReconciler) deleteExternalResources(ctx context.Context, hellowordspec *hellov1.Helloworld, namespace string) error {
	//
	// delete any external resources associated with the cronJob
	//
	// Ensure that delete implementation is idempotent and safe to invoke
	// multiple times for same object.
	podsList := &corev1.PodList{}
	err := r.Client.List(ctx, podsList, &client.ListOptions{Namespace: namespace})
	fmt.Println("PodList Items are....")
	fmt.Println(podsList.Items)
	fmt.Println("PodList Items are....END")
	if err != nil {
		return err
	}
	//pod := []corev1.Pod{}
	for _, v := range podsList.Items {
		if hellowordspec.Spec.PodName == v.Name {
			err = r.Delete(ctx, &v)
			if err != nil {
				return err
			} else {
				fmt.Printf("Pod %s is deleted", v.ObjectMeta.Name)
				return nil
			}
		}

	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *HelloworldReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hellov1.Helloworld{}).
		Complete(r)
}
