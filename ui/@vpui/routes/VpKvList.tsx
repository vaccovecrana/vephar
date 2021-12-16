import * as React from "preact/compat"

import { useContext } from "preact/hooks"
import { lockUi, VpContext, VpStore } from "@vpui/store"
import { rpcKvDel, rpcKvList, rpcKvSetBinary, rpcKvSetText } from "@vpui/rpc"
import { RenderableProps } from "preact"
import { VpKeyPage, VpKeyMeta, ValTxt } from "@vpui/schema"
import VpKeyEditor from "@vpui/components/VpKeyEditor"
import { VpRoute } from "."

type VpKviProps = RenderableProps<{s?: VpStore}>

interface VpKviState {
  editKeyName: boolean
  editKey: VpKeyMeta
  searchKey: string
  page: VpKeyPage
}

const PageSizes = [10, 25, 50, 100]

class VpKvi extends React.Component<VpKviProps, VpKviState> {

  constructor() {
    super()
    this.state = {
      editKeyName: true,
      editKey: undefined,
      searchKey: "",
      page: {
        Keys: [],
        NextKey: "",
        PageSize: PageSizes[0]
      }
    }
  }

  public onKeySearch(prefix: string, offset: string, pageSize: Number) {
    const {dispatch: d} = this.props.s
    lockUi(true, d)
      .then(() => rpcKvList(prefix, offset, pageSize))
      .then(res => {
        this.setState(this.state.page ? {
          page: {
            Keys: [...this.state.page.Keys, ...res.Data.Keys],
            PageSize: res.Data.PageSize,
            NextKey: res.Data.NextKey
          }
        } : {page: res.Data})
      }).then(() => lockUi(false, d))
  }

  public onSearchTermUpdate(e: any) {
    if (e.keyCode === 13) {
      this.setState({...this.state,
        page: {...this.state.page, Keys: []}
      }, () => this.onKeySearch(e.target.value, "", this.state.page.PageSize))
    } else {
      this.setState({...this.state, searchKey: e.target.value})
    }
  }

  public onSelectPageSize(e: any) {
    this.setState({
      page: {...this.state.page,
        PageSize: parseInt(e.target.value)
      }
    })
  }

  public onLoadMore() {
    this.onKeySearch(
      this.state.searchKey,
      this.state.page.NextKey,
      this.state.page.PageSize
    )
  }

  public onCreateKey() {
    if (!this.state.editKey || !this.state.editKeyName) {
      this.setState({...this.state,
        editKeyName: true,
        editKey: {
          keyName: "New-key",
          keyType: ValTxt,
          text: undefined,
          file: undefined
        }
      })
    }
  }

  public onEditKeyCancel() { this.setState({...this.state, editKey: undefined}) }

  public onEditKeyChange(newKey: VpKeyMeta) { this.setState({...this.state, editKey: newKey}) }

  public onEditKeySave(newKey: VpKeyMeta) {
    if (confirm(`Save this key and it's data?\n[${newKey.keyName}]`)) {
      const {dispatch: d} = this.props.s
      lockUi(true, d)
        .then(() => newKey.file ? rpcKvSetBinary(newKey.keyName, newKey.file) : rpcKvSetText(newKey.keyName, newKey.text))
        .then(() => this.onEditKeyChange(undefined))
        .then(() => lockUi(false, d))
        .then(() => alert(`Key saved: [${newKey.keyName}]`))
    }
  }

  public linkOf(route: string, key: string) {
    return`${route}?key=${encodeURIComponent(key)}`
  }

  public onDeleteKey(key: string) {
    if (confirm(`Delete this key and it's data?\n[${key}]`)) {
      const {dispatch: d} = this.props.s
      lockUi(true, d)
        .then(() => rpcKvDel(key))
        .then(() => this.setState({...this.state,
          page: {...this.state.page,
            Keys: this.state.page.Keys.filter(k => k !== key)
          }
        })).then(() => lockUi(false, d))
    }
  }

  public onEditKey(key: string) {
    this.setState({...this.state,
      editKeyName: false,
      editKey: {
        keyName: key,
        keyType: ValTxt,
        file: undefined,
        text: undefined
      }
    })
  }

  public renderCard(key: string) {
    return (
      <div class="card minimal mv8 ph16">
        <div class="row align-center">
          <div class="col xs-6">
            <code>{key}</code>
          </div>
          <div class="col xs-6">
            <div class="txr">
              <a href={this.linkOf(VpRoute.KvGet, key)} target="_blank" ><i class="icono-caretDown" />&nbsp;</a>
              <a onClick={() => this.onEditKey(key)}><i class="icono-hamburger" />&nbsp;</a>
              <a onClick={() => this.onDeleteKey(key)}><i class="icono-cross" /></a>
            </div>
          </div>
        </div>
      </div>
    )
  }

  public render() {
    const {page} = this.state
    return (
      <div>
        <div class="row align-center">
          <div class="col md-10 xs-10">
            <h2>K/V List</h2>
          </div>
          <div class="col md-2 xs-2">
            <div class="txr">
              <button class="addButton" onClick={() => this.onCreateKey()}>
                <i class="icono-plus" />
              </button>
            </div>
          </div>
        </div>

        {this.state.editKey ? (
          <VpKeyEditor allowNameEdit={this.state.editKeyName}
            onSave={nk => this.onEditKeySave(nk)}
            onChange={nk => this.onEditKeyChange(nk)}
            onCancel={() => this.onEditKeyCancel()}
            keyMeta={this.state.editKey} />
        ) : []}

        <div class="row gutter-tiny">
          <div class="col md-8 xs-8">
            <div class="form-group">
              <label class="form-label">Search</label>
              <input class="form-control" value={this.state.searchKey}
                onChange={e => this.onSearchTermUpdate(e)}
                onKeyDown={e => this.onSearchTermUpdate(e)} />
              <span class="form-helper">Key prefix</span>
            </div>
          </div>
          <div class="col md-4 xs-4">
            <div class="form-group">
              <label class="form-label">Items</label>
              <select class="form-control" onChange={e => this.onSelectPageSize(e)}>
                {PageSizes.map(ps => <option>{ps}</option>)}
              </select>
              <span class="form-helper">Items per page</span>
            </div>
          </div>
        </div>
        {page && page.Keys && page.Keys.length !== 0 ? (
          <div class="p8">
            {page.Keys.map(k => this.renderCard(k))}
            {page.NextKey ? (
              <div class="txc pv8">
                <button class="btn primary block" onClick={() => this.onLoadMore()}>
                  Show more
                </button>
              </div>
            ) : []}
          </div>
        ) : []}
      </div>
    )
  }
}

const VpKvList = (props: RenderableProps<{}>) => <VpKvi s={useContext(VpContext)} />
export default VpKvList
