
import * as React from 'react';
import {InputText} from 'primereact/inputtext';
import {Button} from 'primereact/button';
import {MultiSelect} from 'primereact/multiselect';
import {Dialog} from 'primereact/dialog';

import {httpGet, httpPost, httpDelete} from '../utils';

class state {
    id:     number;
    rows:   Array<any>;
    links:  Array<any>;
    allLinks: Array<any>;
    total:  number;
    limit:  number;
    offset: number;
    channels: Array<any>;
    selectedChannelID: number;
    addLink: boolean;
    channelName: string;
    multiChannels: Array<any>;
    showAddLinksDialog: boolean;
}

class CampaignsChannelsComponent extends React.Component<any, state> {
    constructor(props: any) {
        super(props);
        this.state = {
            id: props.id,
            rows: [],
            channels: [],
            links: [],
            allLinks: [],
            total: 0,
            limit: 20,
            offset: 0,
            selectedChannelID: 0,
            addLink: false,
            channelName: '',
            multiChannels: [],
            showAddLinksDialog: false
        };
    }
    loadData(limit: number, offset: number) {
        httpGet(`/api/v1/campaigns/${this.props.id}/freechannels`).then((response: any) => {
            const channels = (response.data.result || []);
            httpGet(`/api/v1/campaigns/${this.state.id}/channels`).then((response: any) => {
                this.setState({
                    limit: limit,
                    offset: offset,
                    channels: channels.map((r: any) => {return {value: r.id, label: r.name}}),
                    rows: (response.data.result ? response.data.result: []),
                });
            })
        })
    }
    loadLinks(channelID: number) {
        httpGet(`/api/v1/campaigns/${this.state.id}/channels/${channelID}/links`).then((response: any) => {
            this.setState({
                selectedChannelID: channelID,
                links: (response.data.result ? response.data.result: []),
            });
        })
    }
    createChannel() {
        httpPost("/api/v1/channels", {name: this.state.channelName}).then((response: any) => {
            const newID = response.result.id;
            this.setState({
                channelName: '',
                channels: this.state.channels.concat([{value: newID, label: this.state.channelName}])
            })
            this.loadData(this.state.limit, this.state.offset);
        })
    }
    addChannels() {
        httpPost(`/api/v1/campaigns/${this.props.id}/channels`, {
            channels: this.state.multiChannels
        }).then(() => {
            this.setState({multiChannels: []});
            this.loadData(this.state.limit, this.state.offset);
        })
    }
    componentDidMount() {
        this.loadData(this.state.limit, this.state.offset);
    }
    deleteChannel(e: any, channelID: number) {
        e.preventDefault();
        httpDelete(`/api/v1/campaigns/${this.state.id}/channels/${channelID}`).then(() => {
            this.loadData(this.state.limit, this.state.offset);
        });
    }
    deleteLinkFromChannel(e: any, linkID: number) {
        e.preventDefault();
        httpDelete(`/api/v1/campaigns/${this.state.id}/channels/${this.state.selectedChannelID}/links/${linkID}`).then(() => {
            this.loadData(this.state.limit, this.state.offset);
        });
    }
    toggleChannel(r: any, value: boolean, index: number) {
        if (value) {
            httpPost(`/api/v1/campaigns/${r.id}/channels/${r.id}/mute`, {}).then(() => {
                this.state.rows[index].is_active = true;
                this.setState({rows: this.state.rows});
            });
        } else {
            httpPost(`/api/v1/campaigns/${r.id}/channels/${r.id}/unmute`, {}).then(() => {
                this.state.rows[index].is_active = false;
                this.setState({rows: this.state.rows});
            });
        }
    }
    toggleLink(r: any, value: boolean, index: number) {
        if (value) {
            httpPost(`/api/v1/campaigns/${r.id}/channels/${r.selectedChannelID}/links/${r.id}/mute`, {}).then(() => {
                this.state.links[index].is_active = true;
                this.setState({links: this.state.links});
            });
        } else {
            httpPost(`/api/v1/campaigns/${r.id}/channels/${r.selectedChannelID}/links/${r.id}/unmute`, {}).then(() => {
                this.state.links[index].is_active = false;
                this.setState({links: this.state.links});
            });
        }
    }
    showAddLinksDialogAction() {
        httpGet(`/api/v1/freelinks/${this.state.selectedChannelID}`).then((response) => {
            this.setState({
                allLinks: response.data.result || [],
                showAddLinksDialog: true
            })
        })
    }
    addLink(linkID: number) {
        httpPost(`/api/v1/campaigns/${this.props.id}/channels/${this.state.selectedChannelID}/links`, {
            linkId: linkID
        }).then(() => {
            httpGet(`/api/v1/freelinks/${this.state.selectedChannelID}`).then((response) => {
                this.setState({
                    allLinks: response.data.result || []
                });
                this.loadLinks(this.state.selectedChannelID);
            })
        })
    }
    render() {

    return (
        <React.Fragment>
            <Dialog header="Add link to a channel" visible={this.state.showAddLinksDialog} style={{width: '50vw'}} modal={true} onHide={() => this.setState({showAddLinksDialog: false})}>
                <ul className="list-group">
                    {this.state.allLinks.length > 0 ?
                     this.state.allLinks.map((l: any) => {
                        return (
                            <li className="list-group-item">
                                <Button label="Add" onClick={(e) => this.addLink(l.id)}/>
                                <div className="long-link"><strong>{l.long}</strong></div>
                                <span className="short-link">{l.short}</span>
                            </li>
                        )
                    }): <span>No links to add</span>}
                </ul>

            </Dialog>

            <div className="row">
                <div className="col-6">
                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Add channels</h3>
                        </div>
                        <div className="card-body">
                            <div>
                                <MultiSelect 
                                value={this.state.multiChannels} 
                                options={this.state.channels} 
                                onChange={(e) => this.setState({multiChannels: e.value})} />
                            </div>
                            <div>
                                <Button label="Add selected" icon="pi pi-plus" className="p-button-warning"
                                onClick={() => this.addChannels()}/>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            <div className="row">
                <div className="col-6">
                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Create channel</h3>
                        </div>
                        <div className="card-body">
                            <div className="p-inputgroup">
                                <InputText placeholder="Campaign name"
                                    value={this.state.channelName} 
                                    onChange={(e: any) => this.setState({channelName: e.target.value})}
                                />
                                <Button icon="pi pi-plus" className="p-button-warning"
                                    onClick={() => this.createChannel()}/>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            <div className="row">
                <div className="col-6">
                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Channels</h3>
                        </div>
                        <div className="card-body">
                            <ul className="list-group">
                                {this.state.rows.map((l: any) => {
                                    return (
                                        <li onClick={(e) => this.loadLinks(l.id)} className="list-group-item">
                                            <div><strong>{l.name}</strong></div>
                                            <i className="badge">
                                                {l.count}
                                            </i>
                                            <button className="btn btn-primary" 
                                                onClick={() => this.showAddLinksDialogAction()}>
                                                    Add Links
                                            </button>
                                            <button className="btn btn-danger" onClick={(e) => this.deleteChannel(e, l.id)}>Remove</button>
                                        </li>
                                    )
                                })}
                            </ul>
                        </div>
                    </div>
                </div>
                <div className="col-6">
                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Links</h3>
                        </div>
                        <div className="card-body">
                            <ul className="list-group">
                                {this.state.links.map((l: any) => {
                                    return (
                                        <li onClick={(e) => this.loadLinks(l.id)} className="list-group-item">
                                            <div className="long-link"><strong>{l.long}</strong></div>
                                            <div className="short-link">{l.short}</div>
                                            <i className="badge">
                                                {l.count}
                                            </i>
                                            <button className="btn btn-danger" onClick={(e) => this.deleteLinkFromChannel(e, l.id)}>Remove</button>
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

export default CampaignsChannelsComponent;