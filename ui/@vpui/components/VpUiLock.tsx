import * as React from "preact/compat"
import {RenderableProps} from "preact"

import { useContext } from "preact/hooks"
import { VpContext, usrMsgClear } from "@vpui/store"

const VpUiLock = (props: RenderableProps<{}>) => {
  const {dispatch: d, state} = useContext(VpContext)
  if (state.lastMessage) {
    alert(JSON.stringify(state.lastMessage, null, 2))
    usrMsgClear(d)
  }
  return (
    <div>
      {props.children}
      {state.uiLocked ? <div class="uiLock" /> : []}
    </div>
  )
}

export default VpUiLock