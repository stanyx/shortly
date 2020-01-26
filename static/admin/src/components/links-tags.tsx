import * as React from 'react';
import {httpPost, httpGet, httpDelete} from '../utils';

class LinkTagsState {
    tag: string;
    tags: Array<any>;
}

class LinkTagsComponent extends React.Component<any, LinkTagsState> {
    constructor(props: any) {
        super(props);
        this.state = {tag: '', tags: []};
    }
    loadTags() {
        httpGet(`/api/v1/users/links?linkID=${this.props.linkID}`).then((response) => {
            this.setState({tags: response.data.result[0].tags || []});
        });
    }
    componentDidMount() {
        this.loadTags();
    }
    submit(e: any) {
        e.preventDefault();
        httpPost('/api/v1/tags/create', {
            linkID: Number(this.props.linkID),
            tag:    this.state.tag,
        }).then(() => {
            this.loadTags();
        })
    }
    deleteTag(tag: string) {
        httpDelete(`/api/v1/tags/${this.props.linkID}/${tag}`).then(() => {
            this.loadTags();
        })
    }
    render() {
        return (
            <React.Fragment>
            <div className="row">
                <div className="col-md-6">
                    <div className="card card-warning">
                        <div className="card-header">
                            <h3 className="card-title">Add tag</h3>
                        </div>
                        <div className="card-body">
                            <form role="form">
                                <div className="row">
                                    <div className="col-sm-6">
                                        <div className="form-group">
                                            <label>Tag</label>
                                            <input value={this.state.tag}
                                            onChange={(e)=>this.setState({tag: e.target.value})}
                                            className="form-control" />
                                        </div>
                                    </div>
                                </div>
                            </form>
                            <button className="btn btn-danger" onClick={(e) => this.submit(e)}>Submit</button>
                        </div>
                    </div>
                </div>
            </div>
            <div className="row">
                <div className="col-md-6">
                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Tags</h3>
                        </div>
                        <div className="card-body">
                            <ul className="list-group">
                                {this.state.tags.map((t) => {
                                    return (
                                        <li className="list-group-item">
                                            {t}
                                            <button className="btn btn-danger float-right" onClick={(e) => this.deleteTag(t)}>
                                                Delete
                                            </button>
                                        </li>
                                    )
                                })}
                            </ul>  
                        </div>
                    </div>
                </div>
            </div>
        </React.Fragment>
        )
    }
}

export default LinkTagsComponent;