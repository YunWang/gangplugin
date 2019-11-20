package predicate

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type NamespacePredicate struct {
	//predicate.Funcs
}

func (p *NamespacePredicate) Create(createEvent event.CreateEvent) bool {
	if createEvent.Meta.GetNamespace() == "kube-system" {
		return false
	}
	return true
}
func (p *NamespacePredicate) Update(updateEvent event.UpdateEvent) bool {
	if updateEvent.MetaNew.GetNamespace() == "kube-system" {
		return false
	}
	return true
}
func (p *NamespacePredicate) Delete(deleteEvent event.DeleteEvent) bool {
	if deleteEvent.Meta.GetNamespace() == "kube-system" {
		return false
	}
	return true
}
func (p *NamespacePredicate) Generic(genericEvent event.GenericEvent) bool {
	if genericEvent.Meta.GetNamespace() == "kube-system" {
		return false
	}
	return true
}
