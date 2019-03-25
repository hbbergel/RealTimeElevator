package queue

import "../bcast"
import "../elevio"	
import "../types"
import "../fsm"
import "math"
import "fmt"

type ElevQueue struct {
	QueueSystem [4][4]int
	CabCall [4]int
	HallCall [4][2]int
	ID string
}




func Distributor(localID string, assignedOrder <-chan types.Order, localOrder chan<- types.Button) {
	netSend := make(chan types.Order)
	netRecv := make(chan types.Order)
	go bcast.Transmitter(15002, netSend)
	go bcast.Receiver(15002, netRecv)


	for{
		select{
		case a := <- assignedOrder:
			netSend <- a

			if a.AssignedTo == localID {
				localOrder <- types.Button{Floor:a.Floor, Type:int(a.Button)}
			}

			fmt.Printf("Local assigned: %+v\n", a)
		case a := <- netRecv:
			if a.AssignedTo == localID {
				localOrder <- types.Button{Floor:a.Floor, Type:int(a.Button)}
			}

			fmt.Printf("Received: %+v\n", a)
		}
	}

}

func Assigner(localID string, buttonPressed <-chan elevio.ButtonEvent, allStates <-chan map[string]types.ElevState, peerList <-chan []string, assignedOrder chan types.Order){
	var peers []string
	var states map[string]types.ElevState
	
	for{
		select{
		case peers = <- peerList:
		
		case states = <- allStates:

		case a := <- buttonPressed:
			aliveStates := make(map[string]types.ElevState)

			for _, id := range(peers) {
				if state, ok := states[id]; ok {
					aliveStates[id] = state
				}
			}
			bestID := findBest(a, aliveStates, localID)
			fmt.Println(bestID)
			 
			b := types.Order{a.Floor, a.Button, bestID}
			fmt.Printf("Assigned order: %+v\n", b)
			assignedOrder <- b

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
        
    case types.MOVING:
        duration += travelTime/2
        state.Floor += convertDirToInt(state)
        
    case types.DOOR_OPEN:
        duration -= doorOpenTime/2
    }


    for {
        if(fsm.ShouldStop(state)){
            fsm.ClearAtCurrentFloor(state)
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






