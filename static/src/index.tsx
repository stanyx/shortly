import * as React from 'react';
import * as ReactDOM from 'react-dom';

import {Shortener} from './components/shortener';

(function() {
    window.onload = function() {
        let urlPath = window.location.pathname;
        console.log("ROUTE DEBUG", urlPath);
        if (urlPath == "/dashboard.html") {
            //ReactDOM.render(<HeaderComponent/>, document.querySelector('#header'));
            //ReactDOM.render(<DashboardComponent/>, document.querySelector('#dashboard'));
            //ReactDOM.render(<SidebarComponent/>, document.querySelector('#dashboard'));
        } else if (urlPath == '/') {
            ReactDOM.render(<Shortener/>, document.querySelector("#shortener-form"));
        } else {
            //ReactDOM.render(<SliderComponent/>, document.querySelector('#slider'));
            //ReactDOM.render(<BillingPlanSelectComponent/>, document.querySelector('#slider'));
        }
    }
})();