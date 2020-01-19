import * as React from 'react';
import * as ReactDOM from 'react-dom';
import {BrowserRouter as Router, Switch, Route} from 'react-router-dom';
import {httpGet} from './utils';

import NavLink from './components/navmenu';
import LinksTable from './components/links-table';
import CreateLinkComponent from './components/create-link';
import RolesTable from './components/roles-table';
import CreateRoleComponent from './components/create-role';

const App = () => {

    return (
        <Router>

            <nav className="main-header navbar navbar-expand navbar-white navbar-light">
            
                <ul className="navbar-nav">
                  <li className="nav-item">
                    <a className="nav-link" data-widget="pushmenu" href="#"><i className="fas fa-bars"></i></a>
                  </li>
                  <li className="nav-item d-none d-sm-inline-block">
                    <a href="/admin" className="nav-link">Dashboard</a>
                  </li>
                  <li className="nav-item d-none d-sm-inline-block">
                    <a href="/admin" className="nav-link">Campaigns</a>
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
                <a href="index3.html" className="brand-link">
                    <img src="../../static/assets/img/logo.png" alt="" className="brand-image" />
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
                        <Route path="/links" exact>
                            <LinksTable />
                        </Route>
                        <Route path="/links/create">
                            <CreateLinkComponent/>
                        </Route>
                        <Route path="/roles" exact>
                            <RolesTable/>
                        </Route>
                        <Route path="/roles/create">
                            <CreateRoleComponent/>
                        </Route>
                        <Route path="/users">
                            <span>Users</span>
                        </Route>
                        <Route path="/groups">
                            <span>Groups</span>
                        </Route>
                        <Route path="/webhooks">
                            <span>Webhooks</span>
                        </Route>
                        <Route path="/" exact>
                            <span>Others</span>
                        </Route>
                    </Switch>
                </section>

            </div>

            <footer className="main-footer">
                <strong>Copyright &copy; 2014-2019 <a href="/">Shortly</a>.</strong>
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