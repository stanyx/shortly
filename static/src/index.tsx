import * as React from 'react';
import * as ReactDOM from 'react-dom';
import {BrowserRouter as Router, Switch, Route, useLocation} from 'react-router-dom';

import {Shortener} from './components/shortener';
import HeaderComponent from './components/header';
import RegisterComponent from './components/register';
import LoginComponent from './components/login';
import FooterComponent from './components/footer';
import axios, { AxiosResponse } from 'axios';



const App = () => {

    const useQuery = () => new URLSearchParams(window.location.search);

    const getComponent = () => {
        const action = useQuery().get('action');
        if (action == 'register') {
            return <RegisterComponent/>;
        } else if (action == 'login') {
            return <LoginComponent/>;
        } else {
            return null;
        }
    }

    return (
        <div>
            <HeaderComponent/>
            <Router>
                <Switch>
                    <Route path="/">
                        <div>
                            {getComponent()}
                        </div>
                    </Route>
                </Switch>
            </Router>
            <FooterComponent />
        </div>
    );
}

(function() {

    const httpGet = (url: string): Promise<AxiosResponse> => {
        return axios.get(url, {
            headers: {
                'x-access-token': localStorage.getItem('token')
            }
        });
    }

    const httpPost = (url: string, body: any): Promise<AxiosResponse> => {
        return axios.post(url, body, {
            headers: {
                'x-access-token': localStorage.getItem('token')
            }
        });
    }

    window.onload = function() {
        let urlPath = window.location.pathname;
        if (urlPath == '/' || urlPath == '/static/') {
            httpGet('/api/v1/user').then((response) => {
                console.log(response.data);
            })
            ReactDOM.render(<Shortener/>, document.querySelector("#shortener-form"));
        } else {
            ReactDOM.render(<App/>, document.querySelector("#app"));
        }
    }
})();