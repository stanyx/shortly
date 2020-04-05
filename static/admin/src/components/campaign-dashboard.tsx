
import * as React from 'react';
import {Accordion, AccordionTab} from 'primereact/accordion';
import {Chart} from 'primereact/chart';
import {httpGet} from '../utils';

class state {
    activeIndex: number;
    selectedCampaign: number;
    selectedLink: number;
    campaigns: Array<any>;
    linksData: any;
    linkStatistics: any;
}

class CampaignsChannelsComponent extends React.Component<any, state> {
    constructor(props: any) {
        super(props);
        this.state = {
            activeIndex: 0,
            selectedCampaign: 0,
            selectedLink: 0,
            campaigns: [],
            linksData: {links: [], data: {}},
            linkStatistics: {referers: {}, locations: {}, clicks: {}}
        };
    }
    loadCampaigns() {
        httpGet('/api/v1/campaigns').then((response) => {
            const campaigns = response.data.result || {};
            this.setState({
                campaigns: campaigns
            })
        })
    }

    componentDidMount() {
        this.loadCampaigns();
    }

    selectChannel(campaignID: number, channelID: number) {
        httpGet(`/api/v1/campaigns/${campaignID}/channels/${channelID}/clicks`).then((response) => {
            this.setState({
                linksData: (response.data.result || {})
            })
        })
    }

    selectLink(linkID: number) {
        const linkData = this.state.linksData.data[linkID];
        this.setState({
            linkStatistics: {
                referrers: [],
                locations: [],
                clicks:    []
            }
        })
    }

    render() {

    return (
        <React.Fragment>
            <div className="row">
                <div className="col-3">
                    <div className="card">
                        <div className="card-header">
                            <h3 className="card-title">Campaigns</h3>
                        </div>
                        <div className="card-body">
                        <Accordion activeIndex={this.state.activeIndex} onTabChange={(e) => this.setState({activeIndex: e.index})}>
                            {this.state.campaigns.map((campaign) => {
                                return (
                                    <AccordionTab header={campaign.name}>
                                        <ul className="list-group">
                                            {campaign.channels.map((channelData: any) => {
                                                return (
                                                    <li className="list-group-item">
                                                        <a onClick={(e) => this.selectChannel(campaign.id, channelData.channelID)}>{channelData.channelName}</a>
                                                        <span className="badge">
                                                            {channelData.count}
                                                        </span>
                                                    </li>
                                                )
                                            })}
                                        </ul>
                                    </AccordionTab>
                                )
                            })}
                        </Accordion>
                        </div>
                    </div>
                </div>
                <div className="col-9">
                    <ul className="list-group">
                        {this.state.linksData.links.map((linkData: any) => {
                            return (
                                <li className="list-group-item">
                                    <a onClick={(e) => this.selectLink(linkData.id, linkData.channelID)}>
                                        {linkData.shortUrl}
                                    </a>
                                    <span className="badge">
                                        {linkData.count}
                                    </span>
                                </li>
                            )
                        })}
                    </ul>
                </div>
            </div>
            <div className="row">
                <div className="card">
                    <div className="card-header">
                        <h3 className="card-title">Link Data</h3>
                    </div>
                    <div className="card-body">
                        <div className="row">
                            <div className="col-12">
                                <Chart type="bar" height="300" options={{title: {
                                    display: true,
                                    text: 'Clicks',
                                    fontSize: 16
                                }, maintainAspectRatio: false}} data={this.state.linkStatistics.clicks}>

                                </Chart>
                            </div>
                        </div>
                        <div className="row">
                            <div className="col-6">
                                <Chart type="pie" options={{title: {
                                    display: true,
                                    text: 'Referrers',
                                    fontSize: 16
                                }}} data={this.state.linkStatistics.referrers}></Chart>
                            </div>
                            <div className="col-6">
                                <Chart type="pie" options={{title: {
                                    display: true,
                                    text: 'Locations',
                                    fontSize: 16
                                }}} data={this.state.linkStatistics.locations}></Chart>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </React.Fragment>
        )
    }
}

export default CampaignsChannelsComponent;