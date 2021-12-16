import * as React from "preact/compat"

import { useContext } from "preact/hooks"
import { hit, lockUi, VpContext, VpStore } from "@vpui/store"
import { rpcRaftStatus } from "@vpui/rpc"
import { RenderableProps } from "preact"
import { VpRoute } from "."

type VpRftProps = RenderableProps<{s?: VpStore}>

class VpRftStatus extends React.Component<VpRftProps> {

  public componentDidMount(): void {
    const {dispatch: d} = this.props.s
    lockUi(true, d)
    .then(() => rpcRaftStatus())
    .then(res => hit({type: VpRoute.RaftStatus, payload: res.Data}, d))
    .then(() => lockUi(false, d))
  }

  private renderKv(lb: string, val: any) {
    return (
      <tr>
        <td>{lb}</td>
        <td>
          <code>{val}</code>
        </td>
      </tr>
    )
  }

  public render() {
    const rft = this.props.s.state.raftStats
    return (
      <div>
        <h2>Raft Status</h2>
        <div class="box">
          <table class="table interactive">
            <thead>
              <tr>
                <th>Param</th>
                <th>Value</th>
              </tr>
            </thead>
            <tbody>
              {this.renderKv("Applied index", rft.applied_index)}
              {this.renderKv("Commit index", rft.commit_index)}
              {this.renderKv("Last contact", rft.last_contact)}
              {this.renderKv("Latest configuration", rft.latest_configuration)}
              {this.renderKv("Peers", rft.num_peers)}
              {this.renderKv("Protocol version", rft.protocol_version)}
              {this.renderKv("State", rft.state)}
              {this.renderKv("Term", rft.term)}
            </tbody>
          </table>
        </div>
      </div>
    )
  }
}

const VpRaftStatus = (props: VpRftProps) => <VpRftStatus s={useContext(VpContext)} />
export default VpRaftStatus
