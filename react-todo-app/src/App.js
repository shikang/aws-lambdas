import React, { Component } from 'react';
import {BrowserRouter as Router, Route} from 'react-router-dom';
import Header from './components/layout/Header'
import Todos from './components/Todos';
import AddTodo from './components/AddTodo';
import About from './components/pages/About';
import axios from 'axios';
//import uuid from 'uuid'

import './App.css';
import Axios from 'axios';

class App extends Component {
    state = {
        todos: [
            //{
            //    //id: uuid.v4(),
            //    id: 1,
            //    title: 'Take out the trash',
            //    completed: false
            //},
            //{
            //    //id: uuid.v4(),
            //    id: 2,
            //    title: 'Dinner with wife',
            //    completed: true
            //},
            //{
            //    //id: uuid.v4(),
            //    id: 3,
            //    title: 'Meeting with boss',
            //    completed: false
            //}
        ]
    }

    componentDidMount() {
        //axios.get(`${'https://cors-anywhere.herokuapp.com/'}https://8482ao82ce.execute-api.ap-southeast-1.amazonaws.com/dev/todos?completed=any&limit=10`)
        //    .then(res => this.setState({todos: res.data}));

        axios.get('https://8482ao82ce.execute-api.ap-southeast-1.amazonaws.com/dev/todos?completed=any&limit=10')
            .then(res => this.setState({todos: res.data}));
    }

    toggleComplete = (id) => {
        var val = false;
        this.state.todos.map(todo => {
            if (todo.id === id) {
                val = !todo.completed;
            }
            return todo;
        });

        axios.post('https://8482ao82ce.execute-api.ap-southeast-1.amazonaws.com/dev/todos/update',
            {
                id: id,
                completed: val
            })
            .then(res => {
                if (res.data.success)
                {
                    this.setState({todos: this.state.todos.map(todo => {
                        if (todo.id === id) {
                            todo.completed = val;
                        }
                        return todo;
                    })});
                }
            });
    }

    deleteTodo = (id) => {
        axios.post('https://8482ao82ce.execute-api.ap-southeast-1.amazonaws.com/dev/todos/delete',
            {
                id: id,
            })
            .then(res => {
                if (res.data.success)
                {
                    this.setState({todos: [...this.state.todos.filter(todo => todo.id !== id)]});
                }
            });
    }

    addTodo =  (title) => {
        axios.put('https://8482ao82ce.execute-api.ap-southeast-1.amazonaws.com/dev/todos/add',
            {
                title: title,
                completed: false
            })
            .then(res => this.setState({todos: [...this.state.todos, res.data]}));
        
    }

    render() {
        return (
            <Router>
                <div className="App">
                    <div className="container">
                        <Header />
                        <Route exact path="/" render={props => (
                            <React.Fragment>
                                <AddTodo addTodo={this.addTodo}/>
                                <Todos todos={this.state.todos} toggleComplete={this.toggleComplete} deleteTodo={this.deleteTodo}/>
                            </React.Fragment>
                        )} />
                        <Route path="/about" component={About} />
                    </div>
                </div>
            </Router>
        );
    }
}

export default App;
