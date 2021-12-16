import VpRaftStatus from "@vpui/routes/VpRaftStatus"
import VpKvList from "@vpui/routes/VpKvList"

export const enum VpRoute {
  RaftStatus = "/raft/status",
  KvList = "/kv/list",
  KvGet = "/kv/get",
  KvSet = "/kv/set",
  KvDel = "/kv/del",
  Ui = "/ui",
  UiKvList = "/ui/kv/list"
}

export { VpRaftStatus, VpKvList }
