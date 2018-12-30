import json

from wiki_api.category import Category, get_pages, get_category, get_subcategories, CategoryJSONEncoder

from flask import Flask, jsonify
app = Flask(__name__)

@app.route('/')
def index():
    return app.send_static_file('index.html')

@app.route("/category/<name>")
def category(name):
    data = get_category(Category(name=name))
    return jsonify(data)

@app.route("/category/<name>/subcategories")
def subcategories(name):
    data = get_subcategories(Category(name=name))
    return jsonify(data)

@app.route("/category/<name>/pages")
def pages(name):
    data = get_pages(Category(name=name))
    return jsonify(data)