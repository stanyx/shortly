
import * as React from 'react';
import {SelectButton} from 'primereact/selectbutton';

import {httpGet} from '../utils';

class state {
    selectedCampaign: number;
    campaigns: Array<any>;
}

class CampaignsChannelsComponent extends React.Component<any, state> {
    constructor(props: any) {
        super(props);
        this.state = {
            selectedCampaign: 0,
            campaigns: []
        };
    }
    loadCampaigns() {
        
    }
    loadData(linkID: number) {
        
    }
    componentDidMount() {
        this.loadCampaigns();
    }

    render() {

    return (
        <React.Fragment>
            <div className="row">
                <div className="col-6">
                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Select campaign</h3>
                        </div>
                        <div className="card-body">
                            <div>
                                <SelectButton
                                value={this.state.selectedCampaign} 
                                options={this.state.campaigns} 
                                onChange={(e) => this.setState({selectedCampaign: e.value})} />
                            </div>
                        </div>
                    </div>
                </div>
            </div>
            <div className="row">
                <div className="col-6">
                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title"></h3>
                        </div>
                        <div className="card-body">
                            
                        </div>
                    </div>
                </div>
            </div>
        </React.Fragment>
        )
    }
}

export default CampaignsChannelsComponent;