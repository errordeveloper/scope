import React from 'react';
import PureRenderMixin from 'react-addons-pure-render-mixin';
import reactMixin from 'react-mixin';

export default class ShowMore extends React.Component {

  constructor(props, context) {
    super(props, context);
    this.handleClick = this.handleClick.bind(this);
  }

  handleClick(ev) {
    ev.preventDefault();
    this.props.handleClick();
  }

  render() {
    const { collection, notShown, expanded, hideNumber } = this.props;
    const showLimitAction = collection && (expanded || notShown > 0);
    const limitActionText = !hideNumber && !expanded && notShown > 0 ? `+${notShown}` : '';
    const limitActionIcon = !expanded && notShown > 0 ? 'fa fa-caret-down' : 'fa fa-caret-up';

    if (!showLimitAction) {
      return <span />;
    }
    return (
      <div className="show-more" onClick={this.handleClick}>
        {limitActionText} <span className={`show-more-icon ${limitActionIcon}`} />
      </div>
    );
  }
}

reactMixin.onClass(ShowMore, PureRenderMixin);
