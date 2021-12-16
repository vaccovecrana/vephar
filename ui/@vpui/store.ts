import * as React from "preact/compat"
import { Context, createContext } from "preact"

import { VpRoute } from "@vpui/routes"
import { VpRaftStats } from "@vpui/schema"

export interface VpState {
  uiLocked: boolean
  lastMessage: any
  raftStats: VpRaftStats
}

export type VpDispatch = (action: VpAction) => void

export interface VpStore {
  state: VpState
  dispatch: VpDispatch
}

export type VpAction =
  | {type: "lockUi", payload: boolean}
  | {type: "usrMsg", payload: string}
  | {type: "usrMsgClear"}
  | {type: VpRoute.RaftStatus, payload: any}

export const hit = (act: VpAction, d: VpDispatch): Promise<void> => {
  d(act)
  return Promise.resolve()
}

export const lockUi = (locked: boolean, d: VpDispatch) => hit({type: "lockUi", payload: locked}, d)
export const usrInfo = (payload: string, d: VpDispatch) => hit({type: "usrMsg", payload}, d)
export const usrError = (payload: string, d: VpDispatch) => hit({type: "usrMsg", payload}, d)
export const usrMsgClear = (d: VpDispatch) => hit({type: "usrMsgClear"}, d)

// export const appListUpdate = (payload: AlAppSummaryDto[], d: VpDispatch) => hit({type: AlUiRoute.V1_APPLICATION_LIST, payload}, d)
// export const appReportUpdate = (payload: AlReport, d: VpDispatch) => hit({type: AlUiRoute.V1_APPLICATION_REPORT, payload}, d)

export const vpReducer: React.Reducer<VpState, VpAction> = (state0: VpState, action: VpAction): VpState => {
  switch (action.type) {
    case "usrMsg": return {...state0, lastMessage: action.payload}
    case "usrMsgClear": return {...state0, lastMessage: undefined}
    case "lockUi": return {...state0, uiLocked: action.payload}
    case VpRoute.RaftStatus: return {...state0, raftStats: action.payload}
  }
}

export const initialState: VpState = {
  lastMessage: undefined,
  uiLocked: false,
  raftStats: {} as VpRaftStats
}

export const VpContext: Context<VpStore> = createContext({
  state: initialState, dispatch: () => {}
})
