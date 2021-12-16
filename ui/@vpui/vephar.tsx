import "nord-ui"
import "nord-ui/dist/dark-theme.css"
import "icono"

import "../res/ui-lock.css"
import "../res/main.scss"

import * as React from "preact/compat"
import * as ReactDOM from "preact/compat"
import { useReducer } from "preact/hooks"

import Router from 'preact-router'

import { VpUiLock, VpMenuLeft, VpMenuTop } from "@vpui/components"
import { VpKvList, VpRaftStatus, VpRoute } from "@vpui/routes"
import { initialState, VpContext, vpReducer } from "@vpui/store"
import { VpVersion } from "@vpui/schema"

class VpShell extends React.Component {
  public render() {
    const [state, dispatch] = useReducer(vpReducer, initialState)
    return (
      <VpContext.Provider value={{state, dispatch}}>
        <VpUiLock>
          <div id="app">
            <div class="row">
              <div class="col sm-1 lg-1 xl-1 sm-down-hide">
                <VpMenuLeft />
              </div>
              <div class="col xs-12 sm-12 md-12 md-up-hide">
                <VpMenuTop />
              </div>
              <div class="col xs-12 sm-12 md-11 lg-11 xl-11">
                <div id="appFrame">
                  <Router>
                    <VpRaftStatus path={VpRoute.Ui} />
                    <VpKvList path={VpRoute.UiKvList} />
                  </Router>
                </div>
              </div>
            </div>
            <div class="row">
              <div class="col auto">
                <div class="txc version p16">
                  {VpVersion}
                </div>
              </div>
            </div>
          </div>
        </VpUiLock>
      </VpContext.Provider>
    )
  }
}

ReactDOM.render(<VpShell />, document.getElementById("root"))
