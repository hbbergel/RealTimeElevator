package queue

import "../bcast"
import "../elevio"	
import "../types"
import "../fsm"
import "../peers"
import "math"
import "fmt"
import "time"


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
			if a.Button != 2{
				//netSend <- a

			ticker := time.NewTicker(time.Millisecond)
			
			
			go func() { 
				for{
					select{
					case <- ticker.C:
						netSend <- a
						fmt.Println("Ticker")
						ticker.Stop()
					}
				}
			}()
		}


			if a.AssignedTo == localID || a.Button == 2 {
				localOrder <- types.Button{Floor:a.Floor, Type:int(a.Button)}
				
				elevio.SetButtonLamp(a.Button, a.Floor, true)
			}

			//fmt.Printf("Local assigned: %+v\n", a)
		case a := <- netRecv:
			elevio.SetButtonLamp(a.Button, a.Floor, true)
			//fmt.Println("Motatt bestilling fra nett")
			if a.AssignedTo == localID {
				localOrder <- types.Button{Floor:a.Floor, Type:int(a.Button)}
			}

			
		}
	}

}

func Assigner(localID string, buttonPressed <-chan elevio.ButtonEvent, allStates <-chan map[string]types.ElevState, peerUpdate <-chan peers.PeerUpdate, assignedOrder chan types.Order){
	var peers []string
	var states map[string]types.ElevState

	for{
		select{
		case a := <- peerUpdate:
            peers = a.Peers
		
		case states = <- allStates:

		case a := <- buttonPressed:	

			aliveStates := make(map[string]types.ElevState)
			fmt.Println("alivestates: %+v\n", aliveStates)

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
		c := timeToIdle(state_cpy, btn)
		if c < bestCost {
			bestCost = c
			bestID = id
		}
	}

	return bestID
}


func timeToIdle(state types.ElevState, btn elevio.ButtonEvent) int {
	const travelTime = 2500
	const doorOpenTime = 3000
	duration := 0

    switch state.State {
    case types.IDLE:
        state.Direction = fsm.ChooseDirection(state)
        if(state.Direction == elevio.MD_Stop){
			distance := closestToOrder(btn, state)
			duration += distance
        }
        
	case types.MOVING:
        duration += travelTime/2
		state.Floor += int(state.Direction)

        
	case types.DOOR_OPEN:
		duration -= doorOpenTime/2
    }


    for {
        fmt.Printf("TTI Iter : %+v\n", state)
        if(fsm.ShouldStop(state)){
            fmt.Printf(" stopping at floor : %+v\n", state.Floor)
            state = fsm.ClearAtCurrentFloor(state, nil)
            fmt.Printf(" after clearing: %+v\n", state)
            duration += doorOpenTime
            state.Direction = fsm.ChooseDirection(state)
            fmt.Printf(" new direction: %+v\n", state.Direction)
            if(state.Direction == elevio.MD_Stop){
                return duration
            }
        }
        state.Floor += int(state.Direction)
		duration += travelTime
    }
}


func closestToOrder(btn elevio.ButtonEvent, state types.ElevState) int {
	
	distance := btn.Floor - state.Floor
	if distance < 0 {
		return -distance
	} else {
		return distance
	}
}

func LostPeers(peerUpdateCh <-chan peers.PeerUpdate, allStatesRx <-chan map[string]types.ElevState, newOrder chan<- types.Button) {
	
	var states map[string]types.ElevState
	

	for {
		select{
        case states = <- allStatesRx:
            
		case a := <- peerUpdateCh:
			lost_id := a.Lost
			if len(lost_id) != 0 {
				for _, id := range lost_id {
					orders := states[id].Orders
                    fmt.Printf("Lost Node %v had orders: %+v\n", id, orders)
					for f := 0; f <=3; f++ {
						for b := 0; b <= 1; b++ {
							if orders[f][b] != 0 {
                                fmt.Printf("Sending newOrder %+v\n", types.Button{Floor: f, Type: b})
								newOrder <- types.Button{Floor: f, Type: b}
							}
						
						}
					}
				}
			}
		}
	}
}