package fsm


import "../elevio"
import "../types"
import "runtime"


func ChooseDirection(e types.ElevState)elevio.MotorDirection{ 
    switch e.Direction {
	case elevio.MD_Up:
        if requests_above(e){
            return elevio.MD_Up
        } else if requests_below(e){
            return elevio.MD_Down
        } else {
            return elevio.MD_Stop
        }
    case elevio.MD_Down:
        if requests_below(e){
            return elevio.MD_Down
        } else if requests_above(e){
            return elevio.MD_Up
        } else {
            return elevio.MD_Stop
        }
        
	
    case elevio.MD_Stop:
		if requests_below(e) {
            return elevio.MD_Down
        } else if requests_above(e){
            return elevio.MD_Up
        } else {
            return elevio.MD_Stop
        }
			
    default:
        return elevio.MD_Stop
    }
    
}

func requests_above(e types.ElevState)bool{
    for f := e.Floor+1; f < types.N_FLOORS; f++ {
        for btn := 0; btn < types.N_BUTTONS; btn++ {
            if e.Orders[f][btn] == 1{
                return true
            }
        }
    }
    return false
}


func requests_below(e types.ElevState) bool{
    for f := 0; f < e.Floor; f++{
        for btn := 0; btn < types.N_BUTTONS; btn++{
            if e.Orders[f][btn] == 1{
                return true
            }
        }
    }
    return false
}


func ShouldStop(e types.ElevState) bool {
    switch (e.Direction){
    case elevio.MD_Down:
        if e.Orders[e.Floor][elevio.BT_HallDown] == 1 {
            return true
        } else if e.Orders[e.Floor][elevio.BT_Cab] == 1 {
            return true
        } else if !requests_below(e) {
            return true
        } else {
            return false
        }
    case elevio.MD_Up:
        if e.Orders[e.Floor][elevio.BT_HallUp] == 1 {
            return true
        } else if e.Orders[e.Floor][elevio.BT_Cab] == 1 {
            return true
        } else if !requests_above(e) {
            return true
        } else {
            return false
        }
    
    default:
        return true
    }
}

func ClearAtCurrentFloor(e types.ElevState){ //, onClearedOrder func(btnType int)) {
	for btn := 0; btn <= 2; btn++ {
		if e.Orders[e.Floor][btn] == 2 {
			e.Orders[e.Floor][btn] = 0
            // onClearedOrder(btn)
            runtime.Goexit()
		}
	}
}

func InitElev(){
    elevio.SetDoorOpenLamp(false)
}