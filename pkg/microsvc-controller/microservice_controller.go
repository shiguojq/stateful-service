/*
Copyright 2022.

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

package microsvccontroller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mscv1 "stateful-service/api/v1"
)

// MicroServiceReconciler reconciles a MicroService object
type MicroServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=msc.shiguojq.com,resources=microservices,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=msc.shiguojq.com,resources=microservices/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=msc.shiguojq.com,resources=microservices/finalizers,verbs=update
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MicroService object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *MicroServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	log := r.Log.WithValues("func", req.NamespacedName)

	log.Info("1. start reconcile logic")
	microSvc := &mscv1.MicroService{}

	if err := r.Get(ctx, req.NamespacedName, microSvc); err != nil {
		if errors.IsNotFound(err) {
			log.Info("2.1. MicroService not found, maybe removed")
			return reconcile.Result{}, nil
		}
		log.Error(err, "2.2 error")
		return ctrl.Result{}, err
	}

	log.Info("3. MicroService: " + microSvc.String())

	pod := &v1.Pod{}
	if err := r.Get(ctx, req.NamespacedName, pod); err != nil {
		if errors.IsNotFound(err) {
			log.Info("4. Pod not exist")
			if err = r.createPod(ctx, microSvc); err != nil {
				log.Error(err, "5.1 error")
				return ctrl.Result{}, err
			}
			if err = r.updateStatus(ctx, microSvc); err != nil {
				log.Error(err, "5.2 error")
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		} else {
			log.Error(err, "6. Error")
			return ctrl.Result{}, nil
		}
	}
	return ctrl.Result{}, nil
}

func (r *MicroServiceReconciler) createPod(ctx context.Context, microSvc *mscv1.MicroService) error {
	log := r.Log.WithValues("func", "createPod")
	env := []v1.EnvVar{}
	env = append(env, v1.EnvVar{
		Name:  "RUNNING_PORT",
		Value: fmt.Sprint(microSvc.Spec.Port),
	})
	env = append(env, microSvc.Spec.Config...)

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: microSvc.Namespace,
			Name:      microSvc.Name,
			Labels: map[string]string{
				"app": microSvc.Name,
			},
		},
		Spec: v1.PodSpec{
			/*Affinity: &v1.Affinity{
				PodAffinity: &v1.PodAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
						{
							PodAffinityTerm: v1.PodAffinityTerm{
								LabelSelector: &metav1.LabelSelector{
									MatchExpressions: []metav1.LabelSelectorRequirement{
										{
											Key:      "app",
											Operator: metav1.LabelSelectorOpIn,
											Values: []string{
												microSvc.Spec.Upstream,
												microSvc.Spec.Downstream,
											},
										},
									},
								},
							},
						},
					},
				},
			},*/
			Containers: []v1.Container{
				{
					Name:            microSvc.Name,
					Image:           microSvc.Spec.Image,
					ImagePullPolicy: "IfNotPresent",
					Ports: []v1.ContainerPort{
						{
							Name:          "rpc",
							Protocol:      v1.ProtocolTCP,
							ContainerPort: microSvc.Spec.Port,
						},
					},
					Env: env,
				},
			},
		},
	}

	log.Info("set reference")
	if err := controllerutil.SetControllerReference(microSvc, pod, r.Scheme); err != nil {
		log.Error(err, "SetControllerReference error")
		return err
	}

	log.Info("start create pod")
	if err := r.Create(ctx, pod); err != nil {
		log.Error(err, "create pod error")
		return err
	}

	log.Info("create pod success")
	return nil
}

func (r *MicroServiceReconciler) updateStatus(ctx context.Context, microSvc *mscv1.MicroService) error {
	log := r.Log.WithValues("func", "updateStatus")

	pod := &v1.Pod{}
	namespacedName := types.NamespacedName{
		Namespace: microSvc.Namespace,
		Name:      microSvc.Name,
	}
	if err := r.Get(ctx, namespacedName, pod); err != nil {
		log.Error(err, "get pod error")
		return err
	}

	microSvc.Status.PodName = pod.Name
	if err := r.Update(ctx, microSvc); err != nil {
		log.Error(err, "update micro service error")
		return err
	}
	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MicroServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mscv1.MicroService{}).
		Complete(r)
}
