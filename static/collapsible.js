function tree() {
	const totalWidth = 960;
	const totalHeight = 800;
	const horizontalSpacing = 270;
	var data,
		i = 0,
        duration = 750,
        margin = {top: 40, right: 0, bottom: 40, left: 80},
	    width = totalWidth - margin.left - margin.right,
        height = totalHeight - margin.top - margin.bottom,
        update;

	const diagonal = d3.linkHorizontal().x(d => d.y).y(d => d.x);

    function chart(selection){
        selection.each(function () {
        	height = height - margin.top - margin.bottom;
			width = width - margin.left - margin.right;

			const zoomListener = d3.zoom()
				.scaleExtent([1 / 2, 4])
				.on("zoom", zoomed);

            // append the svg object to the selection
			const svg = d3.select(this)
				.append('svg')
				.call(zoomListener)
                .attr('width', width + margin.left + margin.right)
                .attr('height', height + margin.top + margin.bottom)
				.append('g')
				.attr('transform', `translate(${margin.left}, ${margin.top})`);

			function zoomed() {
				svg.attr("transform", d3.event.transform);
			}

			function dragged(d) {
				d3.select(this)
					.attr("cx", d.x = d3.event.x)
					.attr("cy", d.y = d3.event.y);
			}

            // declares a tree layout and assigns the size of the tree
			const treemap = d3.tree().size([height, width]);

			// assign parent, children, height, depth
			var root = d3.hierarchy(data, d => d.children);
			root.x0 = height / 2; // left edge of the rectangle
			root.y0 = 0; // top edge of the triangle

			function updateChildren(d) {
				if (d.children) {
					d._children = d.children;
					d.children = null;
					return Promise.resolve(d);
				} else if (d._children) {
					d.children = d._children;
					d._children = null;
					return Promise.resolve(d);
				}

				if (!d.children && !d._children) {
					console.log('HERE', d);
					const name = d.data.name.replace(/^Category:/, '');
					return getSubcategoriesPromise(name)
						.then((data) => {
							const hierarch = data.map(c => {
								const parent = d;
								const obj = { name: c.title, children: null };
								const n = d3.hierarchy(obj, x => x.children);
								n.parent = parent;
								n.depth = parent.depth + 1;
								return n;
							});
							d.children = hierarch.length > 0 ? hierarch : null;
							d.data.children = hierarch.map(d => d.data);
							return d;
						});
				}

				return Promise.resolve(d);
			}

			// toggle children on click
			function click(d) {
				updateChildren(d)
					.then((d) => {
						console.log(d);
						update(d);
					});
			}

	        function update(source) {
				// assigns the x and y position for the nodes
				var treeData = treemap(root);
				console.log(treeData);

		        // compute the new tree layout
		        var nodes = treeData.descendants(),
		            links = treeData.descendants().slice(1);

		        // normalise for fixed depth
		        nodes.forEach(d => d.y = d.depth * horizontalSpacing);

		        // ****************** Nodes section ***************************

		        // update the nodes ...
		        var node = svg.selectAll('g.node')
			        .data(nodes, d => d.id || (d.id = ++i));

		        // Enter any new modes at the parent's previous position.
		        var nodeEnter = node.enter().append('g')
			        .attr('class', 'node')
			        .attr('transform', function(d) {
						const x = source.y0 + margin.top;
						const y = source.x0 + margin.left;
			        	return `translate(${x}, ${y})`;
			        })
			        .on('click', click);

		        // add circle for the nodes
		        nodeEnter.append('circle')
			        .attr('class', 'node')
			        .attr('r', 1e-6)
			        .style('fill', d => d._children ? 'node--internal' : 'node--leaf');

		        // add labels for the nodes
		        nodeEnter.append('text')
			        .attr('dy', '.35em')
			        .attr('x', d =>
			        	d.children || d._children ? 0 : 13
			        )
			        .attr('y', d =>
			        	d.children || d._children ? -margin.top : 0
			        )
					.attr('text-anchor', d =>
						d.children || d._children ? 'middle' : 'start'
			        )
			        .text(d => d.data.name);

		        // add number of children to node circle
		        nodeEnter.append('text')
			        .attr('x', -3)
			        .attr('y', 3)
			        .attr('cursor', 'pointer')
			        .style('font-size', '10px')
			        .text((d) => {
			        	if (d.children) {
							return d.children.length;
						}
			        	else if (d._children) {
							return d._children.length;
						}
			        });

		        // UPDATE
		        var nodeUpdate = nodeEnter.merge(node);

		        // transition to the proper position for the node
		        nodeUpdate.transition().duration(duration)
			        .attr('transform', (d) => {
						const x = d.y + margin.top;
						const y = d.x + margin.left;
			        	return `translate(${x}, ${y})`;
			        });

		        // update the node attributes and style
		        nodeUpdate.select('circle.node')
			        .attr('r', 9)
					.style('fill', d => d._children ? 'node--internal' : 'node--leaf')
			        .attr('cursor', 'pointer');

		        // remove any exiting nodes
		        var nodeExit = node.exit()
			        .transition().duration(duration)
			        .attr('transform', (d) => {
						const x = source.y + margin.top;
						const y = source.x + margin.left;
						return `translate(${x},${y})`;
			        })
			        .remove();

		        // on exit reduce the node circles size to 0
		        nodeExit.select('circle')
			        .attr('r', 1e-6);

		        // on exit reduce the opacity of text labels
		        nodeExit.select('text')
			        .style('fill-opacity', 1e-6);

		        // ****************** links section ***************************

		        // update the links
		        var link = svg.selectAll('path.link')
			        .data(links, function(d) { return d.id });

		        // enter any new links at the parent's previous position
		        var linkEnter = link.enter().insert('path', 'g')
			        .attr('class', 'link')
			        .attr('d', function(d) {
			        	var o = {x: source.x0 + margin.left, y: source.y0 + margin.top};
			        	return diagonal({ source: o, target: o });
			        });

		        // UPDATE
		        var linkUpdate = linkEnter.merge(link);

				// transition back to the parent element position
				// d => diagonal({ source: d, target: d.parent })
		        linkUpdate.transition().duration(duration)
					.attr('d', d => diagonal({ source: d, target: d.parent }));

		        // remove any exiting links
		        var linkExit = link.exit()
			        .transition().duration(duration)
			        .attr('d', (d) => {
			        	var o = {x: source.x, y: source.y};
			        	return diagonal({ source: o, target: o });
			        })
			        .remove();

		        // store the old positions for transition
		        nodes.forEach(function(d) {
		        	d.x0 = d.x + margin.left;
		        	d.y0 = d.y + margin.top;
				});


				// creates a curved (diagonal) path from parent to the child nodes
		        function diagonal(point) {
					const s = point.source;
					const d = point.target;
		        	path = 'M ' + (s.y + margin.top) + ' ' + (s.x + margin.left) +
					        'C ' + ((s.y + d.y + (margin.top * 2)) / 2) + ' ' + (s.x + margin.left) +
					        ', ' + ((s.y + d.y + (margin.top * 2)) / 2) + ' ' + (d.x + margin.left) +
					        ', ' + (d.y + margin.top) + ' ' + (d.x + margin.left);
		        	return path;
				}
			}

			function updateData() {
				var root = d3.hierarchy(data, d => d.children);
				root.x0 = height / 2; // left edge of the rectangle
				root.y0 = 0; // top edge of the triangle
			}

	        update(root, data);
        });
	}

	chart.data = function(value) {
    	if (!arguments.length) return data;
    	data = value;
    	if (typeof updateData === 'function') updateData();
    	return chart;
	};

	chart.width = function(value) {
    	if (!arguments.length) return width;
    	width = value;
    	if (typeof updateWidth === 'function') updateWidth();
    	return chart;
	};

    return chart;
}

function getSubcategoriesPromise(category) {
    const defaultCategory = 'Roman usurpers';
    const cat = category || defaultCategory;
    return d3.json(`/category/${cat}/subcategories`);
}

window.addEventListener('load', () => {
    const categoryInput = document.getElementById('category');
    const defaultCategory = 'Roman emperors';
    categoryInput.value = defaultCategory;
    cyGraph(defaultCategory);

	return getSubcategoriesPromise(defaultCategory)
		.then((data) => {
			const category = defaultCategory;
			const children = data.map(d => ({ name: d.title }));
			const root = { name: category, children };
			const chart = tree().data(root);
			d3.select('#d3').call(chart);
		});
});

/*
const addButton = document.getElementById('add-category');
addButton.addEventListener('click', (event) => {
  const category = document.getElementById('category').value;
  cyGraph(category);
  const svg = d3.select("svg");
  svg.selectAll("*").remove();
  init();
  d3Graph(category);
});
*/