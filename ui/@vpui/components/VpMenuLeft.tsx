import * as React from "preact/compat"

import { VpRoute } from "@vpui/routes"
import VpLogo from "./VpLogo"

const VpMenuLeft = () => (
  <div id="menuLeft">
    <div class="txc">
      <div class="logo mv16">
        <VpLogo />
      </div>
      <div class="pv8">
        <a href={VpRoute.Ui}>
          <i class="icono-sitemap" /><br />
          <small>Raft Status</small>
        </a>
      </div>
      <div class="pv8">
        <a href={VpRoute.UiKvList}>
          <i class="icono-list" /><br />
          <small>K/V List</small>
        </a>
      </div>
    </div>
  </div>
)

export default VpMenuLeft
