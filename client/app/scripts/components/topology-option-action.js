import React from 'react';
import PureRenderMixin from 'react-addons-pure-render-mixin';
import reactMixin from 'react-mixin';

import { changeTopologyOption } from '../actions/app-actions';

export default class TopologyOptionAction extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.onClick = this.onClick.bind(this);
  }

  onClick(ev) {
    ev.preventDefault();
    const { optionId, topologyId, item } = this.props;
    changeTopologyOption(optionId, item.get('value'), topologyId);
  }

  render() {
    const { activeValue, item } = this.props;
    const className = activeValue === item.get('value')
      ? 'topology-option-action topology-option-action-selected' : 'topology-option-action';
    return (
      <div className={className} onClick={this.onClick}>
        {item.get('label')}
      </div>
    );
  }
}

reactMixin.onClass(TopologyOptionAction, PureRenderMixin);
