import React from "react";
import { BrowserRouter as Router, Route, Link } from "react-router-dom";
import {LoginForm} from './login.js'
function BasicExample() {
  CheckAuth()
  return (
    <Router>
        <Route exact path="/" component={Home} />
        <Route path="/about" component={About} />
        <Route path="/topics" component={Topics} />
    
    </Router>
  );
}

function Home() {
    if ( localStorage.getItem('login') === "true"){
    return <Dashboard />
    }
    else {
      return <LoginForm />
    }
  
}


function CheckAuth() {
  fetch('http://localhost:8000/s/refresh', {
      method: 'GET',
      credentials: 'include',

  })
      .then((res) => {
          if (res.status === 401) {
            localStorage.setItem('login', "false");
          } else {
            localStorage.setItem('login', "true");
          }

      },
          (error) => {

            localStorage.setItem('login', "false");

          }
      )

      console.log("Reached end of function")
}



function Topics({ match }) {
  return (
    <div>
      <h2>Topics</h2>
      <ul>
        <li>
          <Link to={`${match.url}/rendering`}>Rendering with React</Link>
        </li>
        <li>
          <Link to={`${match.url}/components`}>Components</Link>
        </li>
        <li>
          <Link to={`${match.url}/props-v-state`}>Props v. State</Link>
        </li>
      </ul>

      <Route path={`${match.path}/:topicId`} component={Topic} />
      <Route
        exact
        path={match.path}
        render={() => <h3>Please select a topic.</h3>}
      />
    </div>
  );
}

function Topic({ match }) {
  return (
    <div>
      <h3>{match.params.topicId}</h3>
    </div>
  );
}

export default BasicExample;
