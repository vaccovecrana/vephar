import { VpRoute } from "@vpui/routes"
import * as React from "preact/compat"
import VpLogo from "./VpLogo"

const VpMenuTop = () => (
  <div class="pt8">
    <div class="row justify-center align-center">
      <div class="col auto">
        <div class="txc logo">
            <VpLogo />
          </div>
        </div>
      <div class="col auto">
        <div class="txc">
          <div class="pv8">
            <a href={VpRoute.Ui}>
              <i class="icono-sitemap" /><br />
              <small>Raft Status</small>
            </a>
          </div>
        </div>
      </div>
      <div class="col auto">
        <div class="txc">
          <div class="pv8">
            <a href={VpRoute.UiKvList}>
              <i class="icono-list" /><br />
              <small>K/V List</small>
            </a>
          </div>
        </div>
      </div>
    </div>
  </div>
)

export default VpMenuTop
