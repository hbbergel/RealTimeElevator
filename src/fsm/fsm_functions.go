package fsm


import "../elevio"
import "../types"
import "fmt"


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
            if e.Orders[f][btn] != 0 {
                return true
            }
        }
    }
    return false
}


func requests_below(e types.ElevState) bool{
    for f := 0; f < e.Floor; f++{
        for btn := 0; btn < types.N_BUTTONS; btn++{
            if e.Orders[f][btn] != 0 {
                return true
            }
        }
    }
    return false
}


func ShouldStop(e types.ElevState) bool {
    switch (e.Direction){
    case elevio.MD_Down:
        if e.Orders[e.Floor][1] != 0 || e.Orders[e.Floor][2] != 0 {
            return true
        } else if !requests_below(e) {
            return true
        } else if e.Floor == 0  {
            return true
        }else {
            return false
        }
    case elevio.MD_Up:
        if e.Orders[e.Floor][0] != 0 || e.Orders[e.Floor][2] != 0 {
            return true
        } else if !requests_above(e) {
            return true
        } else if e.Floor == 3  {
            return true
        } else {
            return false
        }
    case elevio.MD_Stop:
        if ((e.Orders[e.Floor][0] != 0) || (e.Orders[e.Floor][1] != 0) || (e.Orders[e.Floor][2] != 0)){
            return true
        } else{
            return false
        }
    
    default:
        return true
    }
}


func ClearAtCurrentFloor(e types.ElevState, onClearedOrder func(btnType int)) types.ElevState{
   /* if e.Direction == elevio.MD_Down{
        for btn := 0; btn <= 2; btn += 2 {
            if e.Orders[e.Floor][btn] != 0 {
                e.Orders[e.Floor][btn] = 0
                if onClearedOrder != nil {
                    onClearedOrder(btn)
                }
            }
        }
    } else if e.Direction == elevio.MD_Up {
        for btn := 1; btn <= 2; btn ++{
            if e.Orders[e.Floor][btn] != 0 {
                e.Orders[e.Floor][btn] = 0
                if onClearedOrder != nil {
                    onClearedOrder(btn)
                }
            }
        }
    } else if e.Direction == elevio.MD_Stop {
        for btn := 0; btn <= 2; btn ++{
            if e.Orders[e.Floor][btn] != 0 {
                e.Orders[e.Floor][btn] = 0
                if onClearedOrder != nil {
                    onClearedOrder(btn)
                }
            }
        }
    }
    */
    for btn := 0; btn <= 2; btn ++{
        if e.Orders[e.Floor][btn] != 0 {
            e.Orders[e.Floor][btn] = 0
            if onClearedOrder != nil {
                onClearedOrder(btn)
            }
        }
    }
    fmt.Printf("Matrix in fsm,\n\t%+v\n", e.Orders)
    return e
}

