export const VpVersion = "v0.5.0"

export interface VpRpcResponse<T> {
  Error: string
  Data: T
}

export interface VpRaftStats {
  applied_index: Number
  commit_index: Number
  last_contact: string

  latest_configuration: string

  protocol_version: Number
  protocol_version_max: Number
  protocol_version_min: Number

  num_peers: Number
  state: string
  term: Number
}

export interface VpKeyPage {
	Keys: string[]
	NextKey: string
	PageSize: Number
}

export interface VpKeyMeta {
  keyName: string
  keyType: string
  text: string
  file: File
}

export const ValTxt = "Text";
export const ValBin = "Binary"
export const VTypes = [ValTxt, ValBin]
