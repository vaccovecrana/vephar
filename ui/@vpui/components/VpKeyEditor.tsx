import { ValTxt, VpKeyMeta, VTypes } from "@vpui/schema";
import * as React from "preact/compat"

interface VkeProps {
  keyMeta: VpKeyMeta
  allowNameEdit: boolean
  onCancel: () => void
  onChange: (km: VpKeyMeta) => void
  onSave: (km: VpKeyMeta) => void
}

export default class VpKeyEditor extends React.Component<VkeProps> {

  public onKeyTermUpdate(e: any) {
    const kn = e.target.value
    if (!kn) {
      this.props.onCancel()
    } else {
      this.props.onChange({...this.props.keyMeta, keyName: e.target.value})
    }
  }

  public onSelectContentType(e: any) {
    const kt = e.target.value
    const st = {...this.props.keyMeta, keyType: kt}
    if (kt === ValTxt) {
      st.file = undefined
    } else {
      st.text = undefined
    }
    this.props.onChange(st)
  }

  public onSelectFile(e: any) {
    this.props.onChange({...this.props.keyMeta, file: e.target.files[0], text: undefined})
  }

  public onTextValueChange(e: any) {
    this.props.onChange({...this.props.keyMeta, text: e.target.value, file: undefined})
  }

  public render() {
    const {keyMeta: km} = this.props
    const valid = km.file || km.text
    return (
      <div>
        <div class="row gutter-tiny">
          <div class="col auto">
            <div class="form-group">
              <label class="form-label">Key</label>
              <input class="form-control" value={km.keyName}
                onChange={e => this.onKeyTermUpdate(e)}
                disabled={!this.props.allowNameEdit} />
              <span class="form-helper">New key name</span>
            </div>
          </div>
          <div class="col md-4 xs-4">
            <div class="form-group">
              <label class="form-label">Type</label>
              <select class="form-control" onChange={e => this.onSelectContentType(e)}>
                {VTypes.map(ps => <option>{ps}</option>)}
              </select>
              <span class="form-helper">Content Type</span>
            </div>
          </div>
        </div>
        <div class="row gutter-tiny">
          <div class="col auto">
            {km.keyType === ValTxt ? (
              <div class="form-group">
                <label class="form-label">Text</label>
                <textarea class="form-control value-input"
                  value={km.text} rows={8} placeholder="Text value"
                  onChange={e => this.onTextValueChange(e)} />
              </div>
            ) : (
              <div class="form-group">
                <label class="form-label">File</label>
                <div class="card minimal">
                  <div class="card-body">
                    <div class="txc">
                      <input type="file" onChange={e => this.onSelectFile(e)} />
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
        <div class="txc">
          <button class="btn primary small" disabled={!valid}
            onClick={() => this.props.onSave(km)}>Save</button>
          &nbsp;
          <button class="btn secondary small"
            onClick={() => this.props.onCancel()}>Cancel</button>
        </div>
      </div>
    )
  }

}
