import * as React from 'react';
import {Link} from 'react-router-dom';


const NavItem = (props: any) => {
    return (
        <li className="nav-item">
            <Link to={props.url} className={"nav-link " + props.className}>
            <i className={"nav-icon " + props.icon}></i>
            <p>
                {props.title}
                {props.active ? 
                <i className="right fas fa-angle-left"></i>:null
                }
            </p>
            </Link>
        </li>
    )
}


const NavMenu = () => {

    //TODO - make active class

    const menu = [
        {url: '/', title: 'Home', icon: 'fa fa-home', active: true},
        {url: '/dashboards', title: 'Dashboards', icon: 'fa fa-tachometer-alt'},
        {url: '/links', title: 'Links', icon: 'fa fa-link'},
        {url: '/roles', title: 'Roles', icon: 'fa fa-user-tag'},
        {url: '/users', title: 'Users', icon: 'fa fa-user'},
        {url: '/groups', title: 'Groups', icon: 'fa fa-layer-group'},
        {url: '/webhooks', title: 'Webhooks', icon: 'fa fa-plug'},
        {url: '/profile', title: 'Profile', icon: 'fa fa-cog'},
    ];

    return (
        <nav className="mt-2">
            <ul className="nav nav-pills nav-sidebar flex-column" data-widget="treeview" role="menu" data-accordion="false">
                {menu.map((m) => {
                    return <NavItem url={m.url} title={m.title} icon={m.icon}/>
                })}
            </ul>
        </nav>
    )
}

export default NavMenu;