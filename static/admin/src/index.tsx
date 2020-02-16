import * as React from 'react';
import * as ReactDOM from 'react-dom';
import {BrowserRouter as Router, Switch, Route, useParams} from 'react-router-dom';
import {Elements, StripeProvider} from 'react-stripe-elements';
import {httpGet} from './utils';

import 'primereact/resources/themes/nova-light/theme.css';
import 'primereact/resources/primereact.min.css';
import 'primeicons/primeicons.css';

import NavLink from './components/navmenu';
import LinksTable from './components/links-table';
import LinkTagsComponent from './components/links-tags';
import CreateLinkComponent from './components/create-link';
import RolesTable from './components/roles-table';
import CreateRoleComponent from './components/create-role';
import PermissionTable from './components/permissions-table';
import UserTable from './components/users-table';
import CreateUserComponent from './components/create-user';
import ProfileComponent from './components/profile';
import GroupTable from './components/groups-table';
import CreateGroupComponent from './components/create-group';
import ChangeRoleComponent from './components/change-role';
import WebhookTableComponent from './components/webhooks-table';
import CreateWebhookComponent from './components/create-webhook';
import CampaignsTable from './components/campaigns-table';
import CreateCampaignComponent from './components/create-campaign';
import CampaingChannelsComponent from './components/campaign-channels';
import DashboardManagerComponent from './components/dashboards';
import CreateDashboardComponent from './components/create-dashboard';
import CreateDashboardWidgetComponent from './components/create-dashboard-widget';
import MainDashboardComponent from './components/main-dashboard';
import PaymentFormComponent from './components/payment-form';

const PermissionTableComponent = () => {
    let {roleID} = useParams();
    return (
        <PermissionTable roleID={roleID}/>
    );
};

const ChangeRoleComponentWrapper = () => {
    let {userID} = useParams();
    return (
        <ChangeRoleComponent userID={userID}/>
    )
}

const LinkTagsComponentWrapper = () => {
    let {linkID} = useParams();
    return (
        <LinkTagsComponent linkID={linkID}/>
    )
}

const CreateDashboardWidgetComponentWrapper = () => {
    let {dashboardID, posX, posY, maxSpan} = useParams();
    return (
        <CreateDashboardWidgetComponent dashboardID={dashboardID} posX={posX} posY={posY} maxSpan={maxSpan}/>
    )
}

const CampaignChannelsComponentWrapper = () => {
    let {id} = useParams();
    return (
        <CampaingChannelsComponent id={id}/>
    )
}

const PaymentFormComponentWrapper = () => {
    let {planID} = useParams();
    return (
        <StripeProvider apiKey='pk_test_KZTRmzOB7llevD5VHsSVWyev00wfWHxSYy'>
            <Elements>
                <PaymentFormComponent planID={planID}/>
            </Elements>
        </StripeProvider>
    )
}

const App = () => {

    return (
        <Router>

            <nav className="main-header navbar navbar-expand navbar-white navbar-light">
            
                <ul className="navbar-nav">
                  <li className="nav-item">
                    <a className="nav-link" data-widget="pushmenu" href="#"><i className="fas fa-bars"></i></a>
                  </li>
                </ul>
            
                <form className="form-inline ml-3">
                  <div className="input-group input-group-sm">
                    <input className="form-control form-control-navbar" type="search" placeholder="Search" aria-label="Search"/>
                    <div className="input-group-append">
                      <button className="btn btn-navbar" type="submit">
                        <i className="fas fa-search"></i>
                      </button>
                    </div>
                  </div>
                </form>
            </nav>

            <aside className="main-sidebar sidebar-dark-primary elevation-4">
                <a href="/admin" className="brand-link">
                    <i className="brand-image ion ion-analytics"></i>
                    <span className="brand-text font-weight-light">Shortly</span>
                </a>

                <div className="sidebar">
                    <NavLink/>
                </div>
            </aside>

            <div className="content-wrapper">

                <div className="content-header">
                    <div className="container-fluid">
                        <div className="row mb-2">
                        <div className="col-sm-6">
                            <h1 className="m-0 text-dark">Dashboard</h1>
                        </div>
                        <div className="col-sm-6">
                            <ol className="breadcrumb float-sm-right">
                            <li className="breadcrumb-item"><a href="#">Home</a></li>
                            <li className="breadcrumb-item active">Dashboard</li>
                            </ol>
                        </div>
                        </div>
                    </div>
                </div>

                <section className="content">
                    <Switch>
                        <Route path="/" exact>
                            <MainDashboardComponent />
                        </Route>
                        <Route path="/admin" exact>
                            <MainDashboardComponent />
                        </Route>
                        <Route path="/links" exact>
                            <LinksTable />
                        </Route>
                        <Route path="/links/create" exact>
                            <CreateLinkComponent/>
                        </Route>
                        <Route path="/links/:linkID/tags">
                            <LinkTagsComponentWrapper/>
                        </Route>
                        <Route path="/roles" exact>
                            <RolesTable/>
                        </Route>
                        <Route path="/roles/create">
                            <CreateRoleComponent/>
                        </Route>
                        <Route path="/users/:userID/change_role">
                            <ChangeRoleComponentWrapper/>
                        </Route>
                        <Route path="/permissions/:roleID" exact>
                            <PermissionTableComponent/>
                        </Route>
                        <Route path="/users" exact>
                            <UserTable />
                        </Route>
                        <Route path="/users/create" exact>
                            <CreateUserComponent />
                        </Route>
                        <Route path="/groups" exact>
                            <GroupTable />
                        </Route>
                        <Route path="/groups/create" exact>
                            <CreateGroupComponent />
                        </Route>
                        <Route path="/webhooks" exact>
                            <WebhookTableComponent />
                        </Route>
                        <Route path="/webhooks/create">
                            <CreateWebhookComponent />
                        </Route>
                        <Route path="/campaigns" exact>
                            <CampaignsTable />
                        </Route>
                        <Route path="/campaigns/create" exact>
                            <CreateCampaignComponent />
                        </Route>
                        <Route path="/campaigns/:id/channels" exact>
                            <CampaignChannelsComponentWrapper />
                        </Route>
                        <Route path="/profile" exact>
                            <ProfileComponent />
                        </Route>
                        <Route path="/dashboards" exact>
                            <DashboardManagerComponent />
                        </Route>
                        <Route path="/dashboards/create" exact>
                            <CreateDashboardComponent />
                        </Route>
                        <Route path="/dashboards/widgets/:dashboardID/add/:posX/:posY/:maxSpan">
                            <CreateDashboardWidgetComponentWrapper />
                        </Route>
                        <Route path="/billing/:planID/upgrade">
                            <PaymentFormComponentWrapper></PaymentFormComponentWrapper>
                        </Route>
                        <Route path="/" exact>
                            <span>Others</span>
                        </Route>
                    </Switch>
                </section>

            </div>

            <footer className="main-footer">
                <strong>Copyright &copy; 2014-2020 <a href="/">Shortly</a>.</strong>
                All rights reserved.
                <div className="float-right d-none d-sm-inline-block">
                <b>Version</b> 1.0.0
                </div>
            </footer>

            <aside className="control-sidebar control-sidebar-dark">

            </aside>
        </Router>
    );
}

(function() {

    window.onload = function() {
        httpGet('/api/v1/user').then((response: any) => {
            ReactDOM.render(<App/>, document.querySelector("#app"));
        });
    }
})();