package main

import (
	"strings"
)

type Room struct {
	correctOrderOfSurfaces []string // это нужно так как мапа не упорядочена - упорядочивание поверхностей
	objectsInRoom          map[string][]Object
	neighborRooms          []string
	openedDoor             map[string]bool
	goTextMessage          func() string
	viewTextMessage        func() string
}

type Object interface { // какая-то вещь в помещении
	Take() (bool, string) // подобрать
	GetName() string
}

type ItemInInventory struct {
	nameItem string
	useFor   []string // названия объектов, для которых можно применить данный элемент инвентаря
}

type Inventory struct {
	nameInv        string
	objInInventory []ItemInInventory
}

type Player struct {
	curRoom   string
	inventory *Inventory // для рюкзака
}

var rooms map[string]Room = make(map[string]Room)

var player Player

func (i *Inventory) Take() (bool, string) {
	player.inventory = i
	result := "вы надели: " + i.nameInv
	return true, result
}

func (i *Inventory) GetName() string { return i.nameInv }

func (i ItemInInventory) Take() (bool, string) {
	var result string
	if player.inventory != nil {
		player.inventory.objInInventory = append(player.inventory.objInInventory, i)
		result += ("предмет добавлен в инвентарь: " + i.nameItem)
		return true, result
	} else {
		result += "некуда класть"
		return false, result
	}
}

func (i ItemInInventory) GetName() string { return i.nameItem }

func (r Room) ObjInRoom() string {
	var result string
	haveObj := false
	for i, surface := range r.correctOrderOfSurfaces {
		ObjectsOnSurface := r.objectsInRoom[surface]
		if len(ObjectsOnSurface) > 0 {
			haveObj = true
			if i > 0 {
				result += ", "
			}
			result += (surface + ": ")
			for j, object := range ObjectsOnSurface {
				if j != 0 {
					result += ", "
				}
				result += object.GetName()
			}
		}
	}
	if !haveObj {
		result += "пустая комната"
	}
	return result
}

func (r Room) WhereWeGoing() string {
	var result string
	result += "можно пройти - "
	for i, place := range r.neighborRooms {
		if i != 0 {
			result += ", "
		}
		result += place
	}
	return result
}

func (p Player) View() string { // осмотреться
	var result string
	result += rooms[p.curRoom].viewTextMessage()
	return result
}

func (p *Player) Go(in string) string { // идти
	var result string
	present := false
	for _, neighborRoom := range rooms[p.curRoom].neighborRooms {
		if neighborRoom == in {
			present = true
		}
	}
	switch {
	case present:
		if rooms[p.curRoom].openedDoor[in] {
			if in == "домой" {
				p.curRoom = "коридор"
			} else {
				p.curRoom = in
			}
			result += rooms[p.curRoom].goTextMessage()
			result += rooms[p.curRoom].WhereWeGoing()
		} else {
			result += "дверь закрыта"
		} // дверь закрыта
	case func(in string) bool {
		for i := range rooms {
			if i == in || i == "домой" {
				return true
			}
		}
		return false
	}(in): // такое помещение есть, но оно не соседнее
		result += ("нет пути в " + in)
	default:
		result += "нет такой комнаты"
	}
	return result
}

func (p *Player) Take(in string) string { // взять, надеть
	var result string
	for str, sliceObjects := range rooms[p.curRoom].objectsInRoom {
		for i, object := range sliceObjects {
			if object.GetName() == in { // Объект есть
				flag, buf := object.Take()
				result += buf
				if flag { // если есть инвентарь
					rooms[p.curRoom].objectsInRoom[str][i] = rooms[p.curRoom].objectsInRoom[str][len(sliceObjects)-1]
					rooms[p.curRoom].objectsInRoom[str] = rooms[p.curRoom].objectsInRoom[str][:len(sliceObjects)-1]
				}
				return result
			}
		}
	}
	return result + "нет такого"
}

func (p *Player) Use(what string, where string) string { // применить
	var result string
	if p.inventory == nil {
		result += ("нет предмета в инвентаре - " + what)
	} else {
		flag := false
		var item ItemInInventory
		for _, val := range p.inventory.objInInventory {
			if val.nameItem == what {
				flag = true
				item = val
				break
			}
		}
		switch {
		case !flag:
			result += ("нет предмета в инвентаре - " + what)
		case (func(where string) bool { // если истина, то можем применить
			for _, str := range item.useFor {
				if str == where {
					return true
				}
			}
			return false
		})(where):
			if what == "ключи" && where == "дверь" {
				allOpened := true
				for closedRoom, val := range rooms[p.curRoom].openedDoor {
					if !val {
						rooms[p.curRoom].openedDoor[closedRoom] = true
						rooms[closedRoom].openedDoor[p.curRoom] = true
						allOpened = false
						result += "дверь открыта"
						break
					}
				}
				if allOpened {
					result += "дверь открыта"
				}
			}
		default:
			result += "не к чему применить"
		}
	}
	return result
}

func initGame() {
	player.inventory = nil
	player.curRoom = "кухня"

	var r Room // комната
	slice := []string{"коридор"}
	r.neighborRooms = slice
	sliceUseForCurItem := []string{"дверь"} // то, к каким объектам можно применить конкретный элемент
	i1 := ItemInInventory{"ключи", sliceUseForCurItem}
	i2 := ItemInInventory{"конспекты", nil}
	sliceOfItem := []Object{i1, i2}
	r.objectsInRoom = make(map[string][]Object)
	r.objectsInRoom["на столе"] = sliceOfItem
	slice = []string{"на столе", "на стуле"}
	r.correctOrderOfSurfaces = slice
	inv := Inventory{"рюкзак", nil}
	sliceOfItem = []Object{&inv}
	r.objectsInRoom["на стуле"] = sliceOfItem
	r.openedDoor = make(map[string]bool)
	r.openedDoor["коридор"] = true
	r.goTextMessage = func() string { return "ты в своей комнате. " }
	defaultFuncPrintForViewInRoom := func() string {
		return (rooms[player.curRoom].ObjInRoom() + ". " + rooms[player.curRoom].WhereWeGoing())
	}
	r.viewTextMessage = defaultFuncPrintForViewInRoom
	rooms["комната"] = r

	var r1 Room // кухня
	slice = []string{"коридор"}
	r1.neighborRooms = slice
	sliceUseForCurItem = []string{"кружка"} // пример применения объекта, не описанный тестами
	i3 := ItemInInventory{"чай", sliceUseForCurItem}
	r1.objectsInRoom = make(map[string][]Object)
	r1.objectsInRoom["на столе"] = []Object{i3}
	slice = []string{"на столе"}
	r1.correctOrderOfSurfaces = slice
	r1.openedDoor = make(map[string]bool)
	r1.openedDoor["коридор"] = true
	r1.goTextMessage = func() string { return "кухня, ничего интересного. " }
	r1.viewTextMessage = func() string {
		var result string
		result += ("ты находишься на кухне, " + rooms[player.curRoom].ObjInRoom() + ", надо ")
		switch {
		case player.inventory == nil:
			result += "собрать рюкзак и идти в универ. "
		case (func(inv Inventory) bool {
			inventoryReady := 0
			for _, obj := range inv.objInInventory {
				if obj.GetName() == "конспекты" {
					inventoryReady++
				} // для расширяемости
				if inventoryReady == 1 { // всместо 1 можно написать нужное число объектов, необходимых для того чтобы считать рюкзак собранным
					return true
				}
			}
			return false
		})(*player.inventory):
			result += "идти в универ. "
		default:
			result += "собрать рюкзак и идти в универ. "
		}
		result += rooms[player.curRoom].WhereWeGoing()
		return result
	}
	rooms["кухня"] = r1

	var r2 Room // коридор
	slice = []string{"кухня", "комната", "улица"}
	r2.neighborRooms = slice
	r2.openedDoor = make(map[string]bool)
	r2.openedDoor["улица"] = false
	r2.openedDoor["кухня"] = true
	r2.openedDoor["комната"] = true
	r2.goTextMessage = func() string { return "ничего интересного. " }
	r2.viewTextMessage = defaultFuncPrintForViewInRoom
	rooms["коридор"] = r2

	var r3 Room // улица
	slice = []string{"домой"}
	r3.openedDoor = make(map[string]bool)
	r3.openedDoor["домой"] = false
	r3.neighborRooms = slice
	r3.goTextMessage = func() string { return "на улице весна. " }
	r3.viewTextMessage = func() string {
		return "на улице так красиво, что, пожалуй, в универ я не пойду. "
	}
	rooms["улица"] = r3
}

func main() {}

func handleCommand(command string) string {
	strings := strings.Split(command, " ")
	switch strings[0] {
	case "осмотреться":
		return player.View()
	case "идти":
		return player.Go(strings[1])
	case "взять":
		return player.Take(strings[1])
	case "надеть":
		return player.Take(strings[1])
	case "применить":
		return player.Use(strings[1], strings[2])
	default:
		return "неизвестная команда"
	}
}
