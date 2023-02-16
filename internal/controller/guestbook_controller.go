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
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	apiv1 "k8s.io/api/core/v1"
	v1Networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	webappv1 "my.domain/guestbook/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	l "sigs.k8s.io/controller-runtime/pkg/log"
)

// GuestbookReconciler reconciles a Guestbook object
type GuestbookReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=webapp.my.domain,resources=guestbooks,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=webapp.my.domain,resources=guestbooks/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=webapp.my.domain,resources=guestbooks/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Guestbook object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *GuestbookReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := l.FromContext(ctx)

	var guestbook webappv1.Guestbook
	if err := r.Get(ctx, req.NamespacedName, &guestbook); err != nil {
		log.Error(err, "unable to fetch guestbook")
		return ctrl.Result{}, err
	}

	if err := r.ManageDeployment(guestbook, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.ManageService(guestbook, log); err != nil {
		return ctrl.Result{}, err
	}

	if err := r.ManageIngress(guestbook, log); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *GuestbookReconciler) ManageDeployment(guestbook webappv1.Guestbook, log logr.Logger) error {
	dep := NewDeployment(&guestbook)

	existingDep := &appsv1.Deployment{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: dep.Name, Namespace: dep.Namespace}, existingDep); err != nil && errors.IsNotFound(err) {
		log.Info("creating deployment")
		if err := r.Create(context.TODO(), dep); err != nil {
			return err
		}
	} else if err != nil {
		log.Error(err, "deployment already exists")
		return err
	}

	if !equality.Semantic.DeepDerivative(dep.Spec, existingDep.Spec) {
		log.Info("updating deployment")
		if err := r.Update(context.TODO(), dep); err != nil {
			log.Error(err, "error updating deployment")
			return err
		}
	}

	return nil
}

func (r *GuestbookReconciler) ManageService(guestbook webappv1.Guestbook, log logr.Logger) error {
	svc := NewService(&guestbook)

	existingSvc := &apiv1.Service{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, existingSvc); err != nil && errors.IsNotFound(err) {
		log.Info("creating service")
		if err := r.Create(context.TODO(), svc); err != nil {
			return err
		}
	} else if err != nil {
		log.Error(err, "service already exists")
		return err
	}

	return nil
}

func (r *GuestbookReconciler) ManageIngress(guestbook webappv1.Guestbook, log logr.Logger) error {
	ing := NewIngress(&guestbook)

	existingIng := &v1Networking.Ingress{}
	if err := r.Get(context.TODO(), types.NamespacedName{Name: ing.Name, Namespace: ing.Namespace}, existingIng); err != nil && errors.IsNotFound(err) {
		log.Info("creating ingress")
		if err := r.Create(context.TODO(), ing); err != nil {
			return err
		}
	} else if err != nil {
		log.Error(err, "ingress already exists")
		return err
	}

	if !equality.Semantic.DeepDerivative(ing.Spec, existingIng.Spec) {
		log.Info("updating ingress")
		if err := r.Update(context.TODO(), ing); err != nil {
			log.Error(err, "error updating ingress")
			return err
		}
	}

	return nil
}

func NewDeployment(g *webappv1.Guestbook) *appsv1.Deployment {
	replicas := int32(g.Spec.Replicas)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "deployment",
			Namespace: "guestbook-system",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"deployment": g.Name + "-deployment",
				},
			},
			Template: apiv1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"deployment": g.Name + "-deployment",
					},
				},
				Spec: apiv1.PodSpec{
					Containers: []apiv1.Container{
						{
							Name:  g.Name,
							Image: g.Spec.Image,
							Ports: []apiv1.ContainerPort{{
								ContainerPort: 80,
							}},
						},
					},
				},
			},
		},
	}
}

func NewService(g *webappv1.Guestbook) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service",
			Namespace: "guestbook-system",
		},
		Spec: apiv1.ServiceSpec{
			Selector: map[string]string{
				"deployment": g.Name + "-deployment",
			},
			Ports: []apiv1.ServicePort{{
				Port: 80,
			}},
		},
	}
}

func NewIngress(g *webappv1.Guestbook) *v1Networking.Ingress {
	pathTypePrefix := v1Networking.PathTypePrefix

	return &v1Networking.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ingress",
			Namespace: "guestbook-system",
			Annotations: map[string]string{
				"kubernetes.io/ingress.class": "nginx",
				"cert-manager.io/issuer":      "letsencrypt-development",
			},
		},
		Spec: v1Networking.IngressSpec{
			Rules: []v1Networking.IngressRule{
				{
					Host: g.Spec.Host,
					IngressRuleValue: v1Networking.IngressRuleValue{
						HTTP: &v1Networking.HTTPIngressRuleValue{
							Paths: []v1Networking.HTTPIngressPath{
								{
									Path:     "/",
									PathType: &pathTypePrefix,
									Backend: v1Networking.IngressBackend{
										Service: &v1Networking.IngressServiceBackend{
											Name: g.Name + "-service",
											Port: v1Networking.ServiceBackendPort{
												Number: 80,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			TLS: []v1Networking.IngressTLS{
				{
					Hosts:      []string{g.Spec.Host},
					SecretName: "letsencrypt-development",
				},
			},
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *GuestbookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webappv1.Guestbook{}).
		Complete(r)
}
