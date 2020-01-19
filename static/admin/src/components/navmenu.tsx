import * as React from 'react';
import {Link} from 'react-router-dom';


const NavItem = (props: any) => {
    return (
        <li className="nav-item">
            <Link to={props.url} className={props.className}>
            <i className={props.icon}></i>
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
        {url: '/admin', title: 'Dashboard', icon: 'fas fas-tachometer', active: true},
        {url: '/links', title: 'Links', icon: 'fas-link'},
        {url: '/roles', title: 'Roles', icon: 'fas-role'},
        {url: '/users', title: 'Users', icon: 'fas-user'},
        {url: '/groups', title: 'Groups', icon: 'fas-group'},
        {url: '/webhooks', title: 'Webhooks', icon: 'fas-plug'},
        {url: '/profile', title: 'Profile', icon: 'fas-cog'},
    ];

    return (
        <nav className="mt-2">
            <ul className="nav nav-pills nav-sidebar flex-column" data-widget="treeview" role="menu" data-accordion="false">
                {menu.map((m) => {
                    return <NavItem className="nav-item" url={m.url} title={m.title} icon={m.active ? m.icon+"-alt":m.icon}/>
                })}
            </ul>
        </nav>
    )
}

export default NavMenu;