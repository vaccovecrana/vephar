import { VpRoute } from "@vpui/routes"
import { VpKeyPage, VpRaftStats, VpRpcResponse } from "@vpui/schema"

const urlParamsOf = (args: Map<string, string>) => {
  return [...args.entries()]
    .map(([k, v]) => `${encodeURIComponent(k)}=${encodeURIComponent(v)}`)
    .join("&")
}

const buildUrl = (url: string, args: any) => {
  return `${url}?${urlParamsOf(new Map(Object.entries(args)))}`
}

const doBodyRequest = <I, O>(url: string, body: I, init: RequestInit): Promise<O> => {
  const options: any = {...init}
  if (body) {
    options.body = body
  }
  return fetch(url, options)
    .then(response => response.json())
    .then(jData => Promise.resolve(jData as O))
}

export const doJsonIo = <I, O>(url: string, req: I, method: string): Promise<O> => {
  const json = req ? JSON.stringify(req) : undefined
  return doBodyRequest(url, json, {method})
}

export const rpcRaftStatus = (): Promise<VpRpcResponse<VpRaftStats>> => doJsonIo(VpRoute.RaftStatus, undefined, "GET")
export const rpcKvList = (prefix: string, offset: string, pageSize: Number): Promise<VpRpcResponse<VpKeyPage>> =>
  doJsonIo(buildUrl(VpRoute.KvList, {prefix, offset, pageSize}), undefined, "GET")

export const rpcKvSetText = (key: string, t: string): Promise<VpRpcResponse<string>> => {
  const url = buildUrl(VpRoute.KvSet, {key})
  const opts: RequestInit = {method: "POST", headers: {"Content-Type": "text/plain"}}
  return doBodyRequest(url, t, opts)
}

export const rpcKvSetBinary = (key: string, f: File): Promise<VpRpcResponse<string>> => {
  const url = buildUrl(VpRoute.KvSet, {key})
  const opts: RequestInit = {method: "POST", headers: {"Content-Type": f.type}}
  return doBodyRequest(url, f, opts)
}

export const rpcKvDel = (key: string): Promise<VpRpcResponse<string>> => {
  const url = buildUrl(VpRoute.KvDel, {key})
  return doJsonIo(url, undefined, "GET")
}
