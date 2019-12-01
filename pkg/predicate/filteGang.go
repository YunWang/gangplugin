package predicate

import (
	"github.com/YunWang/gangplugin/pkg/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/event"
)

type GangFilter struct {
	//predicate.Funcs
}

func (p *GangFilter) Create(createEvent event.CreateEvent) bool {
	//fmt.Println("Create Gang")
	if createEvent.Meta.GetNamespace()=="default" {
		return true
	}
	return false
}
func (p *GangFilter) Update(updateEvent event.UpdateEvent) bool {
	oldGang:=updateEvent.MetaOld.(*v1.Gang)
	newGang:=updateEvent.MetaNew.(*v1.Gang)
	if updateEvent.MetaNew.GetNamespace()=="default" {
		if oldGang.Status!=newGang.Status {
			//fmt.Println("OldGang:")
			//fmt.Println(oldGang.Status)
			//fmt.Println("NewGang:")
			//fmt.Println(newGang.Status)
			return true
		}
	}
	return false
}
func (p *GangFilter) Delete(deleteEvent event.DeleteEvent) bool {
	//only status change should be passed and deletetimestamp changed should be passed
	//the former indicated pod's status changed
	//the latter indicated pod had been deleted
	if deleteEvent.Meta.GetNamespace()=="default" {
		return true
	}
	return false
}
func (p *GangFilter) Generic(genericEvent event.GenericEvent) bool {
	if genericEvent.Meta.GetNamespace()=="default" {
		return true
	}
	return false
}