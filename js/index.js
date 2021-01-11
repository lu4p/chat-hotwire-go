import './style.css'

// Turbo
const Turbo = require("@hotwired/turbo")
Turbo.connectStreamSource(new WebSocket("ws://" + document.location.host + "/recieve"));


// Stimulus
import { Application } from "stimulus"
import { definitionsFromContext } from "stimulus/webpack-helpers"

const application = Application.start()
const context = require.context("./controllers", true, /\.js$/)
application.load(definitionsFromContext(context))

// Scroll to bottom
var el = document.getElementById("messages")
el.scrollTop = el.scrollHeight

const observerOptions = {
    childList: true,
}

var observer = new MutationObserver(() => el.scrollTop = el.scrollHeight)
observer.observe(el, observerOptions)