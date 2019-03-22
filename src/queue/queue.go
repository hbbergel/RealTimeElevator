package queue

import "../elevio"	
import "../types"
import "../fsm"
import "math"

type ElevQueue struct {
	QueueSystem [4][4]int
	CabCall [4]int
	HallCall [4][2]int
	ID string
}




func Assigner(local_id string, buttonPressed <-chan elevio.ButtonEvent, all_states <-chan map[string]types.ElevState, peerList <-chan []string, assignedOrder_netTx chan<- types.Order){
	for{
		select{
		case a := <- buttonPressed:
			states := <-all_states

			// TODO: filter out dead peers via peerList
			bestID := findBest(a, states, local_id)
			assignedOrder_netTx <- types.Order{a.Floor, a.Button, bestID}		
		}
	}
}

func findBest(btn elevio.ButtonEvent, states map[string]types.ElevState, localID string) string {
	bestCost := math.MaxInt64
	bestID := localID
	for id, state := range(states) {
		state_cpy := state	// copy necessary??
		state_cpy.Orders[btn.Floor][btn.Button] = 2
		c := timeToIdle(state_cpy)
		if c < bestCost {
			bestCost = c
			bestID = id
		}
	}
	return bestID
}


func timeToIdle(state types.ElevState) int {
	const travelTime = 2500
	const doorOpenTime = 3000
    duration := 0
    
    switch state.State {
    case types.IDLE:
        state.Direction = fsm.ChooseDirection(state)
        if(state.Direction == elevio.MD_Stop){
            return duration;
        }
        break
    case types.MOVING:
        duration += travelTime/2
        state.Floor += convertDirToInt(state)
        break
    case types.DOOR_OPEN:
        duration -= doorOpenTime/2
    }


    for {
        if(fsm.ShouldStop(state)){
            fsm.ClearAtCurrentFloor(state, nil)
            duration += doorOpenTime
            state.Direction = fsm.ChooseDirection(state)
            if(state.Direction == elevio.MD_Stop){
                return duration
            }
        }
        state.Floor += convertDirToInt(state)
        duration += travelTime
    }
}

func convertDirToInt(state types.ElevState) int {
	if state.Direction == elevio.MD_Up {
		return 1
	} else if state.Direction == elevio.MD_Down {
		return -1
	} else {
		return 0
	}
}






